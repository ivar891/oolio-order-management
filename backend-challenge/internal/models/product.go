// Package models defines the domain types for the order management API.
package models

// Product represents a food item available for ordering.
type Product struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Price         float64       `json:"price"`
	Currency      string        `json:"currency"`
	Category      string        `json:"category"`
	StockQuantity int           `json:"-"`
	Image         *ProductImage `json:"image,omitempty"`
}

// ProductImage holds responsive image URLs for a product.
type ProductImage struct {
	Thumbnail string `json:"thumbnail"`
	Mobile    string `json:"mobile"`
	Tablet    string `json:"tablet"`
	Desktop   string `json:"desktop"`
}

// PaginatedResponse wraps a paginated list of items with metadata.
type PaginatedResponse struct {
	Data       any `json:"data"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
