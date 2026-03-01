// Package main is the entry point for the order food online API server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oolio-group/order-management/internal/config"
	"github.com/oolio-group/order-management/internal/handler"
	mcpserver "github.com/oolio-group/order-management/internal/mcp"
	"github.com/oolio-group/order-management/internal/middleware"
	"github.com/oolio-group/order-management/internal/repository"
	"github.com/oolio-group/order-management/internal/service"
	"github.com/oolio-group/order-management/internal/telemetry"
)

func main() {
	ctx := context.Background()

	// Load configuration.
	cfg := config.Load()
	log.Printf("Starting Order Food Online API on port %d", cfg.Port)

	// Initialize OpenTelemetry (best-effort, don't fail if Jaeger isn't running).
	_, otelShutdown, err := telemetry.InitTracer(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	if err != nil {
		log.Printf("WARNING: Failed to initialize OpenTelemetry: %v (continuing without tracing)", err)
		otelShutdown = func(_ context.Context) error { return nil }
	}
	defer otelShutdown(ctx) //nolint:errcheck

	// Connect to PostgreSQL.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Run migrations.
	if err := runMigrations(ctx, pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Build bloom filters asynchronously by streaming from S3 URLs.
	// Before filters are ready, promo validation falls through to DB.
	bloomFilters := repository.NewBloomFilterSet()
	bloomFilters.StartLoading(cfg.CouponFileURLs, func(err error) {
		if err != nil {
			log.Printf("WARNING: Bloom filter loading failed: %v (promo validation uses DB only)", err)
		}
	})

	// Create repositories.
	productRepo := repository.NewProductRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	promoRepo := repository.NewPromoRepository(pool, bloomFilters)

	// Create services.
	productSvc := service.NewProductService(productRepo)
	orderSvc := service.NewOrderService(productRepo, orderRepo, promoRepo)

	// Create handlers.
	productHandler := handler.NewProductHandler(productSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)
	healthHandler := handler.NewHealthHandler(pool, bloomFilters)

	// Set up router.
	mux := http.NewServeMux()

	// Health checks (no auth).
	mux.HandleFunc("GET /healthz", healthHandler.Liveness)
	mux.HandleFunc("GET /readyz", healthHandler.Readiness)

	// Product endpoints (no auth).
	mux.HandleFunc("GET /api/product", productHandler.ListProducts)
	mux.HandleFunc("GET /api/product/{productId}", productHandler.GetProduct)

	// Order endpoints (auth required).
	authMiddleware := middleware.Auth(cfg.APIKey)
	mux.Handle("POST /api/order", authMiddleware(http.HandlerFunc(orderHandler.PlaceOrder)))

	// Apply global middleware.
	var httpHandler http.Handler = mux
	httpHandler = middleware.Logging()(httpHandler)
	httpHandler = middleware.Recovery()(httpHandler)
	httpHandler = middleware.CORS()(httpHandler)

	// Create HTTP server.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start MCP server in background.
	mcpSrv := mcpserver.NewServer(productSvc, orderSvc)
	go func() {
		if err := mcpSrv.StartSSE(fmt.Sprintf(":%d", cfg.MCPPort)); err != nil {
			log.Printf("MCP server error: %v", err)
		}
	}()

	// Start HTTP server.
	go func() {
		log.Printf("HTTP server listening on :%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

// runMigrations executes SQL migration files.
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	files := []string{
		"migrations/001_create_schema.up.sql",
		"migrations/002_seed_products.up.sql",
	}

	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", file, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("executing migration %s: %w", file, err)
		}
		log.Printf("Migration applied: %s", file)
	}
	return nil
}
