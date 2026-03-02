package service

import (
	"context"
	"testing"

	"github.com/oolio-group/order-management/internal/models"
	"github.com/oolio-group/order-management/internal/repository"
)

// mockProductRepo is a test double for ProductRepository.
type mockProductRepo struct {
	products map[string]models.Product
}

func (m *mockProductRepo) List(_ context.Context) ([]models.Product, error) {
	result := make([]models.Product, 0, len(m.products))
	for _, p := range m.products {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockProductRepo) GetByID(_ context.Context, id string) (*models.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, models.NewNotFoundError("product not found")
	}
	return &p, nil
}

func (m *mockProductRepo) GetByIDs(_ context.Context, ids []string) ([]models.Product, error) {
	var result []models.Product
	for _, id := range ids {
		if p, ok := m.products[id]; ok {
			result = append(result, p)
		}
	}
	return result, nil
}

// mockOrderRepo is a test double for OrderRepository.
type mockOrderRepo struct {
	lastOrder *models.OrderResponse
}

func (m *mockOrderRepo) Create(_ context.Context, req models.OrderRequest, products []models.Product, total, discounts float64) (*models.OrderResponse, error) {
	currency := "USD"
	if len(products) > 0 {
		currency = products[0].Currency
	}
	resp := &models.OrderResponse{
		ID:         "test-uuid",
		Items:      req.Items,
		Products:   products,
		CouponCode: req.CouponCode,
		Currency:   currency,
		Total:      total,
		Discounts:  discounts,
	}
	m.lastOrder = resp
	return resp, nil
}

// mockPromoRepo is a test double for PromoRepository.
type mockPromoRepo struct {
	validations map[string]*repository.PromoValidation
}

func (m *mockPromoRepo) Validate(_ context.Context, code string) (*repository.PromoValidation, error) {
	v, ok := m.validations[code]
	if !ok {
		return &repository.PromoValidation{Valid: false}, nil
	}
	return v, nil
}

func newTestOrderService() (*OrderService, *mockOrderRepo) {
	products := map[string]models.Product{
		"1": {ID: "1", Name: "Waffle", Price: 6.50, Currency: "USD", StockQuantity: 100},
		"2": {ID: "2", Name: "Crème Brûlée", Price: 7.00, Currency: "USD", StockQuantity: 100},
		"3": {ID: "3", Name: "Macaron", Price: 8.00, Currency: "USD", StockQuantity: 100},
		"5": {ID: "5", Name: "Baklava", Price: 4.00, Currency: "USD", StockQuantity: 100},
	}
	orderRepo := &mockOrderRepo{}
	svc := NewOrderService(
		&mockProductRepo{products: products},
		orderRepo,
		&mockPromoRepo{validations: map[string]*repository.PromoValidation{
			"HAPPYHRS":  {Valid: true, DiscountPercentage: 18.0},
			"BUYGETONE": {Valid: true, DiscountPercentage: 0},
			"VALIDCODE": {Valid: true, DiscountPercentage: 0},
		}},
	)
	return svc, orderRepo
}

func TestPlaceOrderBasic(t *testing.T) {
	svc, _ := newTestOrderService()
	resp, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items: []models.OrderItem{{ProductID: "1", Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 13.0 {
		t.Errorf("expected total 13.0, got %v", resp.Total)
	}
	if resp.Discounts != 0 {
		t.Errorf("expected discounts 0, got %v", resp.Discounts)
	}
	if resp.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", resp.Currency)
	}
}

func TestPlaceOrderWithHappyHrs(t *testing.T) {
	svc, _ := newTestOrderService()
	resp, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items:      []models.OrderItem{{ProductID: "1", Quantity: 2}},
		CouponCode: "HAPPYHRS",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 18% of 13.0 = 2.34 (from DB discount_percentage)
	if resp.Discounts != 2.34 {
		t.Errorf("expected discounts 2.34, got %v", resp.Discounts)
	}
	if resp.Total != 10.66 {
		t.Errorf("expected total 10.66, got %v", resp.Total)
	}
}

func TestPlaceOrderWithBuyGetOne(t *testing.T) {
	svc, _ := newTestOrderService()
	resp, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items: []models.OrderItem{
			{ProductID: "1", Quantity: 1},
			{ProductID: "5", Quantity: 1},
		},
		CouponCode: "BUYGETONE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// discount_percentage=0, so falls back to BuyGetOne strategy: lowest = 4.0
	if resp.Discounts != 4.0 {
		t.Errorf("expected discounts 4.0, got %v", resp.Discounts)
	}
	if resp.Total != 6.5 {
		t.Errorf("expected total 6.5, got %v", resp.Total)
	}
}

func TestPlaceOrderEmptyItems(t *testing.T) {
	svc, _ := newTestOrderService()
	_, err := svc.PlaceOrder(context.Background(), models.OrderRequest{Items: []models.OrderItem{}})
	if err == nil {
		t.Fatal("expected error for empty items")
	}
	apiErr, ok := err.(*models.APIError)
	if !ok || apiErr.Code != "validation" {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestPlaceOrderNilItems(t *testing.T) {
	svc, _ := newTestOrderService()
	_, err := svc.PlaceOrder(context.Background(), models.OrderRequest{})
	if err == nil {
		t.Fatal("expected error for nil items")
	}
}

func TestPlaceOrderNegativeQuantity(t *testing.T) {
	svc, _ := newTestOrderService()
	_, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items: []models.OrderItem{{ProductID: "1", Quantity: -1}},
	})
	if err == nil {
		t.Fatal("expected error for negative quantity")
	}
	apiErr, ok := err.(*models.APIError)
	if !ok || apiErr.Code != "validation" {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestPlaceOrderZeroQuantity(t *testing.T) {
	svc, _ := newTestOrderService()
	_, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items: []models.OrderItem{{ProductID: "1", Quantity: 0}},
	})
	if err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestPlaceOrderEmptyProductID(t *testing.T) {
	svc, _ := newTestOrderService()
	_, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items: []models.OrderItem{{ProductID: "", Quantity: 1}},
	})
	if err == nil {
		t.Fatal("expected error for empty productId")
	}
}

func TestPlaceOrderInvalidProduct(t *testing.T) {
	svc, _ := newTestOrderService()
	_, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items: []models.OrderItem{{ProductID: "999", Quantity: 1}},
	})
	if err == nil {
		t.Fatal("expected error for invalid product")
	}
	apiErr, ok := err.(*models.APIError)
	if !ok || apiErr.Code != "constraint" {
		t.Errorf("expected constraint error, got %v", err)
	}
}

func TestPlaceOrderInvalidPromoCode(t *testing.T) {
	svc, _ := newTestOrderService()
	_, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items:      []models.OrderItem{{ProductID: "1", Quantity: 1}},
		CouponCode: "INVALID1X",
	})
	if err == nil {
		t.Fatal("expected error for invalid promo")
	}
	apiErr, ok := err.(*models.APIError)
	if !ok || apiErr.Code != "validation" {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestPlaceOrderPromoTooShort(t *testing.T) {
	svc, _ := newTestOrderService()
	_, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items:      []models.OrderItem{{ProductID: "1", Quantity: 1}},
		CouponCode: "SHORT",
	})
	if err == nil {
		t.Fatal("expected error for short promo code")
	}
}

func TestPlaceOrderDuplicateProducts(t *testing.T) {
	svc, _ := newTestOrderService()
	resp, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items: []models.OrderItem{
			{ProductID: "1", Quantity: 2},
			{ProductID: "1", Quantity: 3},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 32.5 {
		t.Errorf("expected total 32.5, got %v", resp.Total)
	}
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 merged item, got %d", len(resp.Items))
	}
}

func TestPlaceOrderMultipleItems(t *testing.T) {
	svc, _ := newTestOrderService()
	resp, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items: []models.OrderItem{
			{ProductID: "1", Quantity: 1},
			{ProductID: "2", Quantity: 1},
			{ProductID: "3", Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 21.5 {
		t.Errorf("expected total 21.5, got %v", resp.Total)
	}
	if len(resp.Products) != 3 {
		t.Errorf("expected 3 products, got %d", len(resp.Products))
	}
}

func TestPlaceOrderValidPromoNoStrategy(t *testing.T) {
	svc, _ := newTestOrderService()
	resp, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items:      []models.OrderItem{{ProductID: "1", Quantity: 2}},
		CouponCode: "VALIDCODE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Valid code, 0% discount in DB, no strategy match → no discount
	if resp.Discounts != 0 {
		t.Errorf("expected 0 discounts, got %v", resp.Discounts)
	}
	if resp.Total != 13.0 {
		t.Errorf("expected total 13.0, got %v", resp.Total)
	}
}

func TestPlaceOrderWithDBPercentageDiscount(t *testing.T) {
	products := map[string]models.Product{
		"1": {ID: "1", Name: "Waffle", Price: 6.50, Currency: "USD", StockQuantity: 100},
	}
	orderRepo := &mockOrderRepo{}
	svc := NewOrderService(
		&mockProductRepo{products: products},
		orderRepo,
		&mockPromoRepo{validations: map[string]*repository.PromoValidation{
			"SUMMER25X": {Valid: true, DiscountPercentage: 25.0},
		}},
	)

	resp, err := svc.PlaceOrder(context.Background(), models.OrderRequest{
		Items:      []models.OrderItem{{ProductID: "1", Quantity: 2}},
		CouponCode: "SUMMER25X",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 25% of 13.0 = 3.25
	if resp.Discounts != 3.25 {
		t.Errorf("expected discounts 3.25, got %v", resp.Discounts)
	}
	if resp.Total != 9.75 {
		t.Errorf("expected total 9.75, got %v", resp.Total)
	}
}
