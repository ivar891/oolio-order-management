// Package repository provides data access interfaces and implementations.
package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oolio-group/order-management/internal/models"
)

// ProductRepository defines data access operations for products.
type ProductRepository interface {
	List(ctx context.Context) ([]models.Product, error)
	GetByID(ctx context.Context, id string) (*models.Product, error)
}

type pgProductRepository struct {
	pool *pgxpool.Pool
}

// NewProductRepository creates a new PostgreSQL-backed product repository.
func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &pgProductRepository{pool: pool}
}

// List returns all products.
func (r *pgProductRepository) List(ctx context.Context) ([]models.Product, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, price, currency, category, stock_quantity,
		       image_thumb, image_mobile, image_tablet, image_desktop
		FROM products
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

// GetByID returns a single product by ID.
func (r *pgProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, price, currency, category, stock_quantity,
		       image_thumb, image_mobile, image_tablet, image_desktop
		FROM products
		WHERE id = $1
	`, id)

	var p models.Product
	var thumb, mobile, tablet, desktop *string
	err := row.Scan(&p.ID, &p.Name, &p.Price, &p.Currency, &p.Category, &p.StockQuantity,
		&thumb, &mobile, &tablet, &desktop)
	if err != nil {
		return nil, err
	}

	if thumb != nil || mobile != nil || tablet != nil || desktop != nil {
		p.Image = &models.ProductImage{}
		if thumb != nil {
			p.Image.Thumbnail = *thumb
		}
		if mobile != nil {
			p.Image.Mobile = *mobile
		}
		if tablet != nil {
			p.Image.Tablet = *tablet
		}
		if desktop != nil {
			p.Image.Desktop = *desktop
		}
	}
	return &p, nil
}

// scannable is an interface satisfied by both pgx.Rows and pgx.Row.
type scannable interface {
	Scan(dest ...any) error
}

func scanProduct(s scannable) (models.Product, error) {
	var p models.Product
	var thumb, mobile, tablet, desktop *string
	err := s.Scan(&p.ID, &p.Name, &p.Price, &p.Currency, &p.Category, &p.StockQuantity,
		&thumb, &mobile, &tablet, &desktop)
	if err != nil {
		return p, err
	}

	if thumb != nil || mobile != nil || tablet != nil || desktop != nil {
		p.Image = &models.ProductImage{}
		if thumb != nil {
			p.Image.Thumbnail = *thumb
		}
		if mobile != nil {
			p.Image.Mobile = *mobile
		}
		if tablet != nil {
			p.Image.Tablet = *tablet
		}
		if desktop != nil {
			p.Image.Desktop = *desktop
		}
	}
	return p, nil
}
