package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"

	"github.com/oolio-group/order-management/internal/models"
	"github.com/oolio-group/order-management/internal/repository"
)

var tracer = otel.Tracer("order-service")

// OrderService provides order-related business operations.
type OrderService struct {
	productRepo repository.ProductRepository
	orderRepo   repository.OrderRepository
	promoRepo   repository.PromoRepository
}

// NewOrderService creates a new OrderService.
func NewOrderService(
	productRepo repository.ProductRepository,
	orderRepo repository.OrderRepository,
	promoRepo repository.PromoRepository,
) *OrderService {
	return &OrderService{
		productRepo: productRepo,
		orderRepo:   orderRepo,
		promoRepo:   promoRepo,
	}
}

// PlaceOrder validates and creates an order.
func (s *OrderService) PlaceOrder(ctx context.Context, req models.OrderRequest) (*models.OrderResponse, error) {
	ctx, span := tracer.Start(ctx, "PlaceOrder")
	defer span.End()

	slog.InfoContext(ctx, "Placing new order", slog.Int("item_count", len(req.Items)))

	// 1. Validate items.
	if err := s.validateItems(req.Items); err != nil {
		slog.WarnContext(ctx, "Invalid order items", slog.Any("error", err))
		return nil, err
	}

	// 2. Merge duplicate product IDs.
	req.Items = mergeItems(req.Items)

	// 3. Resolve products and validate promo code in parallel.
	var products []*models.Product
	var validation *repository.PromoValidation
	var discounts float64

	g, gCtx := errgroup.WithContext(ctx)

	// Task A: Resolve products.
	g.Go(func() error {
		_, innerSpan := tracer.Start(gCtx, "ResolveProducts")
		defer innerSpan.End()
		var err error
		products, err = s.resolveProducts(gCtx, req.Items)
		return err
	})

	// Task B: Validate promo code (if provided).
	if req.CouponCode != "" {
		couponCode := strings.TrimSpace(req.CouponCode)
		req.CouponCode = couponCode

		// Validate promo code length.
		if len(couponCode) < 3 || len(couponCode) > 20 {
			return nil, models.NewValidationError(fmt.Sprintf("invalid coupon code: %s", couponCode))
		}

		g.Go(func() error {
			_, innerSpan := tracer.Start(gCtx, "ValidatePromo")
			defer innerSpan.End()
			var err error
			validation, err = s.promoRepo.Validate(gCtx, couponCode)
			if err != nil {
				return fmt.Errorf("validate promo code: %w", err)
			}
			return nil
		})
	}

	// Wait for both tasks to complete.
	if err := g.Wait(); err != nil {
		slog.ErrorContext(ctx, "Parallel validation failed", slog.Any("error", err))
		return nil, err
	}

	// 4. Calculate total.
	itemsWithPrices := buildItemsWithPrices(req.Items, products)
	total := calculateTotal(itemsWithPrices)

	// 5. Apply promo code results.
	if validation != nil && validation.Valid {
		couponCode := req.CouponCode
		// Apply discount: use DB-stored discount percentage if > 0,
		// otherwise fall back to known strategy patterns.
		if validation.DiscountPercentage > 0 {
			discounts = roundToTwoDecimals(total * validation.DiscountPercentage / 100)
		} else if strategy := GetDiscountStrategy(couponCode); strategy != nil {
			if couponCode == "BUYGETONE" {
				var totalQty int
				for _, item := range req.Items {
					totalQty += item.Quantity
				}
				if totalQty < 2 {
					return nil, models.NewValidationError("BUYGETONE requires at least 2 items")
				}
			}
			discounts = strategy.Calculate(itemsWithPrices)
		}
		slog.InfoContext(ctx, "Promo code applied", slog.String("coupon", couponCode), slog.Float64("discount", discounts))
	} else if req.CouponCode != "" {
		// If a code was provided but validation came back invalid.
		slog.WarnContext(ctx, "Invalid coupon code", slog.String("coupon", req.CouponCode))
		return nil, models.NewValidationError(fmt.Sprintf("invalid coupon code: %s", req.CouponCode))
	}

	finalTotal := roundToTwoDecimals(total - discounts)
	if finalTotal < 0 {
		finalTotal = 0
	}

	// 6. Persist order with stock check (atomic).
	resp, err := s.orderRepo.Create(ctx, req, products, finalTotal, discounts)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create order", slog.Any("error", err))
		return nil, err
	}

	slog.InfoContext(ctx, "Order placed successfully", slog.String("order_id", resp.ID))
	return resp, nil
}

// validateItems checks that items are valid.
func (s *OrderService) validateItems(items []models.OrderItem) error {
	if len(items) == 0 {
		return models.NewValidationError("at least one item is required")
	}
	for _, item := range items {
		if strings.TrimSpace(item.ProductID) == "" {
			return models.NewValidationError("productId is required for all items")
		}
		if item.Quantity <= 0 {
			return models.NewValidationError("quantity must be positive")
		}
	}
	return nil
}

// mergeItems combines duplicate product IDs by summing their quantities.
func mergeItems(items []models.OrderItem) []models.OrderItem {
	merged := make(map[string]int)
	order := make([]string, 0)
	for _, item := range items {
		if _, exists := merged[item.ProductID]; !exists {
			order = append(order, item.ProductID)
		}
		merged[item.ProductID] += item.Quantity
	}

	result := make([]models.OrderItem, 0, len(merged))
	for _, pid := range order {
		result = append(result, models.OrderItem{
			ProductID: pid,
			Quantity:  merged[pid],
		})
	}
	return result
}

// resolveProducts fetches products for all items in bulk and validates they exist.
func (s *OrderService) resolveProducts(ctx context.Context, items []models.OrderItem) ([]*models.Product, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// 1. Collect unique product IDs.
	idSet := make(map[string]struct{})
	for _, item := range items {
		idSet[item.ProductID] = struct{}{}
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	// 2. Fetch all products in a single bulk query.
	products, err := s.productRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("resolve products: %w", err)
	}

	// 3. Verify all products exist.
	if len(products) != len(ids) {
		// Figure out which one is missing for a better error message.
		found := make(map[string]struct{})
		for _, p := range products {
			found[p.ID] = struct{}{}
		}
		for _, id := range ids {
			if _, ok := found[id]; !ok {
				return nil, models.NewConstraintError(fmt.Sprintf("invalid product specified: %s", id))
			}
		}
	}

	return products, nil
}

func buildItemsWithPrices(items []models.OrderItem, products []*models.Product) []itemWithPrice {
	priceMap := make(map[string]float64)
	for _, p := range products {
		priceMap[p.ID] = p.Price
	}

	result := make([]itemWithPrice, 0, len(items))
	for _, item := range items {
		result = append(result, itemWithPrice{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     priceMap[item.ProductID],
		})
	}
	return result
}

func calculateTotal(items []itemWithPrice) float64 {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return roundToTwoDecimals(total)
}
