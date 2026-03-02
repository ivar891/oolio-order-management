package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/oolio-group/order-management/internal/config"
	"github.com/oolio-group/order-management/internal/logger"
)

func main() {
	// Initialize logging
	slog.SetDefault(logger.New("development", os.Stdout))

	// Load configuration
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("Failed to ping database", slog.Any("error", err))
		os.Exit(1)
	}

	start := time.Now()
	slog.Info("Starting coupon loading process")

	// 1. Create temporary table for raw data (persistent but short-lived)
	_, err = pool.Exec(ctx, `
		DROP TABLE IF EXISTS raw_promo_codes_loading;
		CREATE TABLE raw_promo_codes_loading (
			code TEXT NOT NULL,
			file_index SMALLINT NOT NULL
		);
	`)
	if err != nil {
		slog.Error("Failed to create temp table", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Exec(ctx, "DROP TABLE IF EXISTS raw_promo_codes_loading;")

	// 2. Load data from S3 URLs into temp table
	for i, url := range cfg.CouponFileURLs {
		if url == "" {
			continue
		}
		fileIndex := int16(i + 1)
		slog.Info("Streaming file", slog.Int("index", i+1), slog.String("url", url))

		if err := streamToTemp(ctx, pool, url, fileIndex); err != nil {
			slog.Error("Failed to stream file", slog.String("url", url), slog.Any("error", err))
			os.Exit(1)
		}
	}

	// 3. Transform and Move valid codes to promo_codes
	slog.Info("Identifying duplicates and moving to promo_codes")
	tag, err := pool.Exec(ctx, `
		INSERT INTO promo_codes (code, file_index, discount_percentage)
		WITH valid_codes AS (
			-- Codes present in at least 2 distinct files
			SELECT code
			FROM raw_promo_codes_loading
			GROUP BY code
			HAVING COUNT(DISTINCT file_index) >= 2
		),
		discounted_codes AS (
			-- Assign discount percentages based on user requirements
			SELECT 
				trc.code,
				trc.file_index,
				CASE 
					WHEN trc.code IN ('HAPPYHRS', 'HAPPYHOURS') THEN 18.0
					WHEN trc.code = 'BUYGETONE' THEN 0.0
					ELSE floor(random() * 10)
				END AS discount_percentage
			FROM raw_promo_codes_loading trc
			JOIN valid_codes vc ON trc.code = vc.code
		)
		SELECT code, file_index, discount_percentage FROM discounted_codes
		ON CONFLICT (code, file_index) DO UPDATE SET discount_percentage = EXCLUDED.discount_percentage;
	`)
	if err != nil {
		slog.Error("Failed to migrate codes", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Coupon loading complete",
		slog.Int64("rows_affected", tag.RowsAffected()),
		slog.Duration("took", time.Since(start)),
	)
}

func streamToTemp(ctx context.Context, pool *pgxpool.Pool, url string, fileIndex int16) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	// Use CopyFrom for high-performance bulk insertion
	scanner := bufio.NewScanner(gz)
	var batch [][]any
	const batchSize = 100000

	for scanner.Scan() {
		code := strings.TrimSpace(scanner.Text())
		if code == "" {
			continue
		}
		batch = append(batch, []any{code, fileIndex})

		if len(batch) >= batchSize {
			if _, err := pool.CopyFrom(ctx, pgx.Identifier{"raw_promo_codes_loading"}, []string{"code", "file_index"}, pgx.CopyFromRows(batch)); err != nil {
				return fmt.Errorf("CopyFrom: %w", err)
			}
			batch = batch[:0]
			slog.Debug("Batch inserted", slog.Int("size", batchSize))
		}
	}

	if len(batch) > 0 {
		if _, err := pool.CopyFrom(ctx, pgx.Identifier{"raw_promo_codes_loading"}, []string{"code", "file_index"}, pgx.CopyFromRows(batch)); err != nil {
			return fmt.Errorf("CopyFrom final batch: %w", err)
		}
	}

	return scanner.Err()
}
