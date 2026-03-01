package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/oolio-group/order-management/internal/models"
	"github.com/oolio-group/order-management/internal/repository"
)

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
	// 1. Validate items.
	if err := s.validateItems(req.Items); err != nil {
		return nil, err
	}

	// 2. Merge duplicate product IDs.
	req.Items = mergeItems(req.Items)

	// 3. Resolve products and validate they exist.
	products, err := s.resolveProducts(ctx, req.Items)
	if err != nil {
		return nil, err
	}

	// 4. Calculate total.
	itemsWithPrices := buildItemsWithPrices(req.Items, products)
	total := calculateTotal(itemsWithPrices)

	// 5. Validate and apply promo code.
	var discounts float64
	if req.CouponCode != "" {
		couponCode := strings.TrimSpace(req.CouponCode)
		req.CouponCode = couponCode

		// Validate promo code length.
		if len(couponCode) < 8 || len(couponCode) > 10 {
			return nil, models.NewValidationError("invalid coupon code")
		}

		// Validate via bloom filter + DB (returns discount percentage).
		validation, err := s.promoRepo.Validate(ctx, couponCode)
		if err != nil {
			return nil, fmt.Errorf("validate promo code: %w", err)
		}
		if !validation.Valid {
			return nil, models.NewValidationError("invalid coupon code")
		}

		// Apply discount: use DB-stored discount percentage if > 0,
		// otherwise fall back to known strategy patterns.
		if validation.DiscountPercentage > 0 {
			discounts = roundToTwoDecimals(total * validation.DiscountPercentage / 100)
		} else if strategy := GetDiscountStrategy(couponCode); strategy != nil {
			discounts = strategy.Calculate(itemsWithPrices)
		}
	}

	finalTotal := roundToTwoDecimals(total - discounts)
	if finalTotal < 0 {
		finalTotal = 0
	}

	// 6. Persist order with stock check (atomic).
	resp, err := s.orderRepo.Create(ctx, req, products, finalTotal, discounts)
	if err != nil {
		return nil, err
	}

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

// resolveProducts fetches products for all items and validates they exist.
func (s *OrderService) resolveProducts(ctx context.Context, items []models.OrderItem) ([]models.Product, error) {
	products := make([]models.Product, 0, len(items))
	for _, item := range items {
		p, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, models.NewConstraintError(fmt.Sprintf("invalid product specified: %s", item.ProductID))
		}
		products = append(products, *p)
	}
	return products, nil
}

func buildItemsWithPrices(items []models.OrderItem, products []models.Product) []itemWithPrice {
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
