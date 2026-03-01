package models

// OrderRequest represents the incoming order placement request.
type OrderRequest struct {
	CouponCode string      `json:"couponCode,omitempty"`
	Items      []OrderItem `json:"items"`
}

// OrderItem represents a single line item in an order.
type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// OrderResponse represents the order confirmation returned to the client.
type OrderResponse struct {
	ID         string      `json:"id"`
	Items      []OrderItem `json:"items"`
	Products   []Product   `json:"products"`
	CouponCode string      `json:"couponCode,omitempty"`
	Currency   string      `json:"currency"`
	Total      float64     `json:"total"`
	Discounts  float64     `json:"discounts"`
}
