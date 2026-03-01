// Package main provides a seed script that runs migrations to populate
// the database with sample products and promo codes.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oolio-group/order-management/internal/config"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	// Run schema migration.
	if err := runSQL(ctx, pool, "migrations/001_create_schema.up.sql"); err != nil {
		log.Fatalf("Schema migration failed: %v", err)
	}
	log.Println("✓ Schema created")

	// Seed products and known promo codes.
	if err := runSQL(ctx, pool, "migrations/002_seed_products.up.sql"); err != nil {
		log.Fatalf("Product seed failed: %v", err)
	}
	log.Println("✓ 9 products + promo codes seeded")

	// Print summary.
	var productCount, promoCount int64
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&productCount)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM promo_codes").Scan(&promoCount)
	log.Printf("Summary: %d products, %d promo code entries", productCount, promoCount)
	log.Println("Note: Bloom filters are loaded from S3 URLs at server startup (not via seed)")
}

func runSQL(ctx context.Context, pool *pgxpool.Pool, file string) error {
	sql, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read %s: %w", file, err)
	}
	_, err = pool.Exec(ctx, string(sql))
	return err
}
