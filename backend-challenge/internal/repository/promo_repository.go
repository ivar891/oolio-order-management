package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PromoValidation holds the result of promo code validation.
type PromoValidation struct {
	Valid              bool
	DiscountPercentage float64
}

// PromoRepository defines operations for promo code validation.
type PromoRepository interface {
	Validate(ctx context.Context, code string) (*PromoValidation, error)
}

type pgPromoRepository struct {
	pool  *pgxpool.Pool
	bloom *BloomFilterSet
}

// NewPromoRepository creates a new PostgreSQL-backed promo repository with bloom filter pre-check.
func NewPromoRepository(pool *pgxpool.Pool, bloom *BloomFilterSet) PromoRepository {
	return &pgPromoRepository{pool: pool, bloom: bloom}
}

// Validate checks if a promo code is valid (exists in at least 2 data files)
// and returns the associated discount percentage.
// Uses bloom filter as a fast pre-check to avoid unnecessary DB queries.
func (r *pgPromoRepository) Validate(ctx context.Context, code string) (*PromoValidation, error) {
	// Pre-check: bloom filter can definitively say "no".
	if r.bloom != nil && !r.bloom.MaybeValid(code) {
		return &PromoValidation{Valid: false}, nil
	}

	// Exact check in DB: must exist in ≥2 files, get the max discount percentage.
	var valid bool
	var discountPct float64
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT file_index) >= 2,
			COALESCE(MAX(discount_percentage), 0)
		FROM promo_codes
		WHERE code = $1
	`, code).Scan(&valid, &discountPct)
	if err != nil {
		return nil, err
	}

	return &PromoValidation{
		Valid:              valid,
		DiscountPercentage: discountPct,
	}, nil
}
