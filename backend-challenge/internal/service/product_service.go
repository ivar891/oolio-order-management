// Package service implements the business logic for the order management API.
package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/oolio-group/order-management/internal/models"
	"github.com/oolio-group/order-management/internal/repository"
)

// ProductService provides product-related business operations.
type ProductService struct {
	repo repository.ProductRepository
}

// NewProductService creates a new ProductService.
func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// ListProducts returns all available products.
func (s *ProductService) ListProducts(ctx context.Context) ([]*models.Product, error) {
	return s.repo.List(ctx)
}

// ListProductsPaginated returns a paginated list of products.
func (s *ProductService) ListProductsPaginated(ctx context.Context, page, limit int) (*models.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	products, total, err := s.repo.ListPaginated(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit // ceiling division

	return &models.PaginatedResponse{
		Data:       products,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetProduct returns a single product by ID.
// Returns a NotFoundError if the product doesn't exist.
func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.NewNotFoundError("product not found")
		}
		return nil, err
	}
	return p, nil
}
