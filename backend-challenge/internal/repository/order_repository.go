package repository

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oolio-group/order-management/internal/models"
)

// OrderRepository defines data access operations for orders.
type OrderRepository interface {
	// Create atomically creates an order, checking and decrementing stock.
	// Returns InsufficientStockError if any product lacks sufficient stock.
	Create(ctx context.Context, req models.OrderRequest, products []models.Product, total, discounts float64) (*models.OrderResponse, error)
}

type pgOrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates a new PostgreSQL-backed order repository.
func NewOrderRepository(pool *pgxpool.Pool) OrderRepository {
	return &pgOrderRepository{pool: pool}
}

// Create places an order atomically with stock checking via SELECT FOR UPDATE.
func (r *pgOrderRepository) Create(ctx context.Context, req models.OrderRequest, products []models.Product, total, discounts float64) (*models.OrderResponse, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Collect and sort product IDs for consistent locking order (deadlock prevention).
	productIDs := make([]string, 0, len(req.Items))
	for _, item := range req.Items {
		productIDs = append(productIDs, item.ProductID)
	}
	sort.Strings(productIDs)

	// Lock product rows and check stock.
	rows, err := tx.Query(ctx, `
		SELECT id, name, stock_quantity
		FROM products
		WHERE id = ANY($1)
		FOR UPDATE
	`, productIDs)
	if err != nil {
		return nil, fmt.Errorf("lock products: %w", err)
	}

	type stockInfo struct {
		name  string
		stock int
	}
	stockMap := make(map[string]stockInfo)
	for rows.Next() {
		var id, name string
		var stock int
		if err := rows.Scan(&id, &name, &stock); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan stock: %w", err)
		}
		stockMap[id] = stockInfo{name: name, stock: stock}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stock rows: %w", err)
	}

	// Verify stock for each item.
	for _, item := range req.Items {
		info, ok := stockMap[item.ProductID]
		if !ok {
			return nil, models.NewConstraintError(fmt.Sprintf("invalid product specified: %s", item.ProductID))
		}
		if info.stock < item.Quantity {
			return nil, models.NewInsufficientStockError(info.name, info.stock, item.Quantity)
		}
	}

	// Decrement stock.
	for _, item := range req.Items {
		_, err := tx.Exec(ctx, `
			UPDATE products SET stock_quantity = stock_quantity - $1, updated_at = NOW()
			WHERE id = $2
		`, item.Quantity, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("update stock: %w", err)
		}
	}

	// Insert order.
	orderID := uuid.New()
	// Determine currency from first product (all products in an order share the same currency).
	currency := "USD"
	if len(products) > 0 {
		currency = products[0].Currency
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, coupon_code, currency, total, discounts)
		VALUES ($1, $2, $3, $4, $5)
	`, orderID, nilIfEmpty(req.CouponCode), currency, total, discounts)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	// Insert order items.
	for _, item := range req.Items {
		// Find price for this product.
		var unitPrice float64
		for _, p := range products {
			if p.ID == item.ProductID {
				unitPrice = p.Price
				break
			}
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, unit_price)
			VALUES ($1, $2, $3, $4)
		`, orderID, item.ProductID, item.Quantity, unitPrice)
		if err != nil {
			return nil, fmt.Errorf("insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &models.OrderResponse{
		ID:         orderID.String(),
		Items:      req.Items,
		Products:   products,
		CouponCode: req.CouponCode,
		Currency:   currency,
		Total:      total,
		Discounts:  discounts,
	}, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
