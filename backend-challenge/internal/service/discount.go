package service

import (
	"math"
)

// DiscountStrategy defines how a discount is calculated.
type DiscountStrategy interface {
	// Calculate returns the discount amount for the given items and their prices.
	Calculate(items []itemWithPrice) float64
}

type itemWithPrice struct {
	ProductID string
	Quantity  int
	Price     float64
}

// HappyHoursStrategy applies 18% discount on the total.
type HappyHoursStrategy struct{}

// Calculate returns 18% of the order total.
func (h *HappyHoursStrategy) Calculate(items []itemWithPrice) float64 {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return roundToTwoDecimals(total * 0.18)
}

// BuyGetOneStrategy makes the lowest-priced item free.
type BuyGetOneStrategy struct{}

// Calculate returns the price of the lowest-priced item.
func (b *BuyGetOneStrategy) Calculate(items []itemWithPrice) float64 {
	if len(items) == 0 {
		return 0
	}
	lowest := items[0].Price
	for _, item := range items[1:] {
		if item.Price < lowest {
			lowest = item.Price
		}
	}
	return roundToTwoDecimals(lowest)
}

// discountStrategies maps coupon codes to their discount strategies.
var discountStrategies = map[string]DiscountStrategy{
	"HAPPYHRS":  &HappyHoursStrategy{},
	"BUYGETONE": &BuyGetOneStrategy{},
}

// GetDiscountStrategy returns the strategy for the given coupon code, or nil.
func GetDiscountStrategy(code string) DiscountStrategy {
	return discountStrategies[code]
}

func roundToTwoDecimals(v float64) float64 {
	return math.Round(v*100) / 100
}
