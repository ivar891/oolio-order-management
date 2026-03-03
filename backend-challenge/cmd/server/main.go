// Package main is the entry point for the order food online API server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oolio-group/order-management/internal/config"
	"github.com/oolio-group/order-management/internal/handler"
	"github.com/oolio-group/order-management/internal/logger"
	"github.com/oolio-group/order-management/internal/middleware"
	"github.com/oolio-group/order-management/internal/repository"
	"github.com/oolio-group/order-management/internal/service"
	"github.com/oolio-group/order-management/internal/telemetry"
)

func main() {
	ctx := context.Background()

	// Load configuration.
	cfg := config.Load()

	// Initialize structured logging.
	appLogger := logger.New(os.Getenv("APP_ENV"), os.Stdout)
	slog.SetDefault(appLogger)

	slog.Info("Starting Order Food Online API", slog.Int("port", cfg.Port))

	// Initialize OpenTelemetry (best-effort, don't fail if Jaeger isn't running).
	_, otelShutdown, err := telemetry.InitTracer(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	if err != nil {
		slog.Warn("Failed to initialize OpenTelemetry", slog.Any("error", err), slog.String("info", "continuing without tracing"))
		otelShutdown = func(_ context.Context) error { return nil }
	}
	defer otelShutdown(ctx) //nolint:errcheck

	// Connect to PostgreSQL.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", slog.Any("error", err), slog.String("hint", "Ensure Docker is running and the database container is started (make docker-up)"))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("Failed to ping database", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Connected to PostgreSQL")

	// Run migrations.
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("Failed to run migrations", slog.Any("error", err))
		os.Exit(1)
	}

	// Build bloom filters asynchronously by streaming from S3 URLs.
	// Before filters are ready, promo validation falls through to DB.
	bloomFilters := repository.NewBloomFilterSet()
	bloomFilters.StartLoading(cfg.CouponFileURLs, func(err error) {
		if err != nil {
			slog.Warn("Bloom filter loading failed", slog.Any("error", err), slog.String("info", "promo validation uses DB only"))
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
	go func() {
		slog.Info("HTTP server listening", slog.Int("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", slog.Any("error", err))
	}
	slog.Info("Server stopped")
}

// runMigrations applies database migrations using golang-migrate.
// It provides version tracking, dirty state detection, advisory locking,
// and rollback support via .down.sql files.
func runMigrations(databaseURL string) error {
	// golang-migrate's pgx5 driver expects the "pgx5://" scheme.
	migrateURL := strings.Replace(databaseURL, "postgres://", "pgx5://", 1)

	m, err := migrate.New("file://migrations", migrateURL)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	slog.Info("Migrations up to date", slog.Uint64("version", uint64(version)), slog.Bool("dirty", dirty))
	return nil
}
