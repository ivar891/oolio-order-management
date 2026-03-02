// Package main provides a seed script that runs migrations to populate
// the database with sample products and promo codes.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oolio-group/order-management/internal/config"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	// Run schema migration.
	if err := runSQL(ctx, pool, "migrations/001_create_schema.up.sql"); err != nil {
		slog.Error("Schema migration failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("✓ Schema created")

	// Seed products and known promo codes.
	if err := runSQL(ctx, pool, "migrations/002_seed_products.up.sql"); err != nil {
		slog.Error("Product seed failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("✓ 9 products + promo codes seeded")

	// Print summary.
	var productCount, promoCount int64
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&productCount)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM promo_codes").Scan(&promoCount)
	slog.Info("Seed summary",
		slog.Int64("products", productCount),
		slog.Int64("promo_codes", promoCount),
	)
	slog.Info("Note: Bloom filters are loaded from S3 URLs at server startup (not via seed)")
}

func runSQL(ctx context.Context, pool *pgxpool.Pool, file string) error {
	sql, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read %s: %w", file, err)
	}
	_, err = pool.Exec(ctx, string(sql))
	return err
}
