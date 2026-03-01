package service

import (
	"testing"
)

func TestHappyHoursDiscount(t *testing.T) {
	strategy := &HappyHoursStrategy{}
	items := []itemWithPrice{
		{ProductID: "1", Quantity: 2, Price: 6.50},
		{ProductID: "2", Quantity: 1, Price: 7.00},
	}
	// Total = 13.0 + 7.0 = 20.0. 18% of 20.0 = 3.6
	discount := strategy.Calculate(items)
	if discount != 3.6 {
		t.Errorf("expected 3.6, got %v", discount)
	}
}

func TestHappyHoursDiscountSingleItem(t *testing.T) {
	strategy := &HappyHoursStrategy{}
	items := []itemWithPrice{
		{ProductID: "1", Quantity: 2, Price: 6.50},
	}
	// Total = 13.0. 18% of 13.0 = 2.34
	discount := strategy.Calculate(items)
	if discount != 2.34 {
		t.Errorf("expected 2.34, got %v", discount)
	}
}

func TestBuyGetOneDiscount(t *testing.T) {
	strategy := &BuyGetOneStrategy{}
	items := []itemWithPrice{
		{ProductID: "1", Quantity: 2, Price: 6.50},
		{ProductID: "2", Quantity: 1, Price: 4.00},
		{ProductID: "3", Quantity: 1, Price: 8.00},
	}
	// Lowest price = 4.00
	discount := strategy.Calculate(items)
	if discount != 4.0 {
		t.Errorf("expected 4.0, got %v", discount)
	}
}

func TestBuyGetOneDiscountEmpty(t *testing.T) {
	strategy := &BuyGetOneStrategy{}
	discount := strategy.Calculate([]itemWithPrice{})
	if discount != 0 {
		t.Errorf("expected 0, got %v", discount)
	}
}

func TestGetDiscountStrategyKnown(t *testing.T) {
	s := GetDiscountStrategy("HAPPYHRS")
	if s == nil {
		t.Error("expected non-nil strategy for HAPPYHRS")
	}
	s = GetDiscountStrategy("BUYGETONE")
	if s == nil {
		t.Error("expected non-nil strategy for BUYGETONE")
	}
}

func TestGetDiscountStrategyUnknown(t *testing.T) {
	s := GetDiscountStrategy("UNKNOWN")
	if s != nil {
		t.Error("expected nil strategy for unknown code")
	}
}

func TestRoundToTwoDecimals(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{1.006, 1.01},
		{1.004, 1.0},
		{0.0, 0.0},
		{100.999, 101.0},
		{2.345, 2.35},
	}
	for _, tt := range tests {
		result := roundToTwoDecimals(tt.input)
		if result != tt.expected {
			t.Errorf("roundToTwoDecimals(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
