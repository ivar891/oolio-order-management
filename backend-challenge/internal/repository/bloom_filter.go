package repository

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
)

const (
	// Each S3 file contains ~102M promo codes (3 files, up to ~306M unique total).
	seenEstimate  = 306_000_000
	seenFPP       = 0.01  // 1% — temporary filter, freed after loading (~350MB)
	validEstimate = 50_000_000
	validFPP      = 0.001 // 0.1% — kept in memory for query-time lookups (~86MB)

	minCodeLen = 8
	maxCodeLen = 10
)

// BloomFilterSet uses a two-phase bloom filter for promo code validation.
//
// Two-phase loading (sequential, no concurrency issues):
//   - Phase 1: Stream file 1 into a "seen" filter (all codes encountered so far).
//   - Phase 2: Stream files 2..N. For each code, if already in "seen",
//     add to "valid" filter (confirmed in ≥2 files). Then add to "seen".
//
// After loading, "seen" is freed (~350MB reclaimed). Only "valid" stays (~86MB).
// MaybeValid does a single bloom filter lookup on "valid".
// False positives fall through to the DB (the authoritative source of truth).
// Before loading completes, MaybeValid returns true (conservative: allows DB fallback).
type BloomFilterSet struct {
	valid *bloom.BloomFilter
	ready atomic.Bool
	mu    sync.RWMutex
}

// NewBloomFilterSet creates an unready bloom filter set.
func NewBloomFilterSet() *BloomFilterSet {
	return &BloomFilterSet{}
}

// StartLoading kicks off background two-phase loading from the given URLs.
// Each URL should point to a gzipped text file with one promo code per line.
// onComplete is called when loading finishes (can be nil).
func (b *BloomFilterSet) StartLoading(urls []string, onComplete func(err error)) {
	go func() {
		err := b.loadFromURLs(urls)
		if err != nil {
			slog.Warn("Bloom filter loading failed", slog.Any("error", err),
				slog.String("info", "promo validation uses DB only"))
		}
		if onComplete != nil {
			onComplete(err)
		}
	}()
}

// IsReady returns true once the valid filter has been loaded.
func (b *BloomFilterSet) IsReady() bool {
	return b.ready.Load()
}

// MaybeValid returns true if the code might exist in ≥2 files.
// Single bloom filter lookup on the pre-computed "valid" filter.
// A false return means the code is definitely NOT in ≥2 files (no false negatives).
// If filters are not yet ready, returns true (conservative: fall through to DB).
func (b *BloomFilterSet) MaybeValid(code string) bool {
	if len(code) < minCodeLen || len(code) > maxCodeLen {
		return false
	}

	if !b.ready.Load() {
		return true
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.valid.TestString(code)
}

func (b *BloomFilterSet) loadFromURLs(urls []string) error {
	if len(urls) < 2 {
		return fmt.Errorf("need at least 2 URLs for ≥2-file intersection")
	}

	start := time.Now()

	// "seen": tracks all codes encountered so far across files.
	// Temporary — freed after loading to reclaim memory.
	seen := bloom.NewWithEstimates(seenEstimate, seenFPP)

	// "valid": codes confirmed present in ≥2 files.
	// Kept in memory permanently for query-time lookups.
	valid := bloom.NewWithEstimates(validEstimate, validFPP)

	// --- Phase 1: Stream first file into "seen" only ---
	slog.Info("Phase 1: loading first file into seen filter",
		slog.String("url", urls[0]))

	count, err := streamGzipURL(urls[0], func(code string) {
		seen.AddString(code)
	})
	if err != nil {
		return fmt.Errorf("file 1 (%s): %w", urls[0], err)
	}
	slog.Info("Phase 1 complete",
		slog.Int64("codes", count),
		slog.Duration("elapsed", time.Since(start)))

	// --- Phase 2: Stream remaining files sequentially ---
	// If a code is already in "seen", it appeared in a prior file → add to "valid".
	// Always add to "seen" so subsequent files detect overlap with this file too.
	var validCount int64
	for i := 1; i < len(urls); i++ {
		url := urls[i]
		if url == "" {
			continue
		}
		slog.Info("Phase 2: processing file against seen filter",
			slog.Int("file", i+1),
			slog.Int("total", len(urls)),
			slog.String("url", url))

		fileCount, err := streamGzipURL(url, func(code string) {
			if seen.TestString(code) {
				valid.AddString(code)
				validCount++
			}
			seen.AddString(code)
		})
		if err != nil {
			return fmt.Errorf("file %d (%s): %w", i+1, url, err)
		}
		slog.Info("File processed",
			slog.Int("file", i+1),
			slog.Int64("codes_in_file", fileCount),
			slog.Int64("valid_so_far", validCount))
	}

	// Atomically swap in the valid filter.
	b.mu.Lock()
	b.valid = valid
	b.mu.Unlock()
	b.ready.Store(true)

	// "seen" goes out of scope and becomes eligible for GC (~350MB freed).
	slog.Info("Two-phase bloom filter loading complete",
		slog.Int64("valid_codes", validCount),
		slog.Duration("total_time", time.Since(start)))

	return nil
}

// streamGzipURL streams a gzipped text file from a URL, decompresses on-the-fly,
// and calls fn for each valid promo code line.
// Memory: only one buffered line at a time (no full-file buffering).
// Returns the number of codes processed.
func streamGzipURL(url string, fn func(code string)) (int64, error) {
	client := &http.Client{Timeout: 10 * time.Minute}

	var resp *http.Response
	var err error

	// Retry with exponential backoff for transient network issues.
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = client.Get(url) //nolint:gosec // URLs are from trusted env config
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		backoff := time.Duration(1<<attempt) * time.Second
		slog.Warn("Retrying bloom filter download",
			slog.String("url", url),
			slog.Int("attempt", attempt+1),
			slog.Any("error", err),
			slog.Duration("backoff", backoff))
		time.Sleep(backoff)
	}
	if err != nil {
		return 0, fmt.Errorf("HTTP GET (after retries): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("gzip reader: %w", err)
	}
	defer gzReader.Close()

	scanner := bufio.NewScanner(bufio.NewReaderSize(gzReader, 128*1024))
	var count int64
	for scanner.Scan() {
		code := strings.TrimSpace(scanner.Text())
		if len(code) >= minCodeLen && len(code) <= maxCodeLen {
			fn(code)
			count++
		}
	}
	if err := scanner.Err(); err != nil && err != io.ErrUnexpectedEOF {
		return count, fmt.Errorf("scanner: %w", err)
	}

	return count, nil
}

// NewEmptyBloomFilterSet returns an immediately-ready empty bloom filter set (for testing).
func NewEmptyBloomFilterSet() *BloomFilterSet {
	b := &BloomFilterSet{
		valid: bloom.NewWithEstimates(100, 0.01),
	}
	b.ready.Store(true)
	return b
}
