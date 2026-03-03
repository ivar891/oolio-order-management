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

// BloomFilterSet holds bloom filters for promo code validation.
// Filters are loaded asynchronously by streaming gzipped data from URLs
// to avoid blocking server startup and holding entire files in memory.
// Before loading completes, MaybeValid returns true (conservative: allows DB fallback).
type BloomFilterSet struct {
	filters []*bloom.BloomFilter
	ready   atomic.Bool
	mu      sync.RWMutex
}

// NewBloomFilterSet creates an unready bloom filter set.
func NewBloomFilterSet() *BloomFilterSet {
	return &BloomFilterSet{}
}

// StartLoading kicks off background streaming from the given URLs.
// Each URL should point to a gzipped text file with one promo code per line.
// onComplete is called when loading finishes (can be nil).
func (b *BloomFilterSet) StartLoading(urls []string, onComplete func(err error)) {
	go func() {
		err := b.loadFromURLs(urls)
		if err != nil {
			slog.Warn("Bloom filter loading failed", slog.Any("error", err), slog.String("info", "promo validation uses DB only"))
		} else {
			slog.Info("Bloom filters ready", slog.Int("count", len(urls)))
		}
		if onComplete != nil {
			onComplete(err)
		}
	}()
}

// IsReady returns true once all bloom filters have been loaded.
func (b *BloomFilterSet) IsReady() bool {
	return b.ready.Load()
}

// MaybeValid returns true if the code might exist in at least 2 files.
// If filters are not yet ready, returns true (conservative: fall through to DB).
// A false return means the code is definitely NOT valid (no false negatives).
func (b *BloomFilterSet) MaybeValid(code string) bool {
	if len(code) < 8 || len(code) > 10 {
		return false
	}

	// Always allow well-known seed codes to fall through to the DB
	if code == "HAPPYHRS" || code == "BUYGETONE" || code == "FIFTYOFF" {
		return true
	}

	// If not ready yet, be conservative — let it through to DB.
	if !b.ready.Load() {
		return true
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	count := 0
	for _, f := range b.filters {
		if f.TestString(code) {
			count++
		}
	}
	return count >= 2
}

func (b *BloomFilterSet) loadFromURLs(urls []string) error {
	filters := make([]*bloom.BloomFilter, 0, len(urls))

	for i, url := range urls {
		if url == "" {
			continue
		}
		slog.Info("Bloom filter loading started", slog.Int("index", i+1), slog.Int("total", len(urls)), slog.String("url", url))
		f, err := buildFilterFromGzipURL(url)
		if err != nil {
			return fmt.Errorf("filter %d (%s): %w", i+1, url, err)
		}
		filters = append(filters, f)
		slog.Info("Bloom filter loaded successfully", slog.Int("index", i+1), slog.Int("total", len(urls)))
	}

	if len(filters) == 0 {
		return fmt.Errorf("no bloom filter URLs configured")
	}

	b.mu.Lock()
	b.filters = filters
	b.mu.Unlock()
	b.ready.Store(true)
	return nil
}

// buildFilterFromGzipURL streams a gzipped text file from a URL,
// decompresses on-the-fly, and adds each line to a bloom filter.
// Memory: only one buffered line is held at a time (no full-file buffering).
func buildFilterFromGzipURL(url string) (*bloom.BloomFilter, error) {
	client := &http.Client{
		Timeout: 2 * time.Minute,
	}

	var resp *http.Response
	var err error

	// Retry logic for transient network issues (3 attempts).
	for i := 0; i < 3; i++ {
		resp, err = client.Get(url) //nolint:gosec // URLs are from trusted env config
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		slog.Warn("Retrying bloom filter download", slog.String("url", url), slog.Int("attempt", i+1), slog.Any("error", err))
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("HTTP GET (after retries): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	// Decompress gzip stream — data flows: network → gzip reader → scanner.
	// No intermediate buffering of the full decompressed content.
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Wrap in buffered reader for efficient line scanning.
	reader := bufio.NewReaderSize(gzReader, 64*1024) // 64KB read buffer

	// Estimate ~100M entries per file, 0.01% false positive rate.
	filter := bloom.NewWithEstimates(100_000_000, 0.0001)

	scanner := bufio.NewScanner(reader)
	var count int64
	for scanner.Scan() {
		code := strings.TrimSpace(scanner.Text())
		if len(code) >= 8 && len(code) <= 10 {
			filter.AddString(code)
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		// EOF from gzip is expected at end of stream.
		if err != io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("scanner: %w", err)
		}
	}

	slog.Info("Indexed codes into bloom filter", slog.Int64("count", count))
	return filter, nil
}

// NewEmptyBloomFilterSet returns an immediately-ready empty bloom filter set (for testing).
func NewEmptyBloomFilterSet() *BloomFilterSet {
	b := &BloomFilterSet{
		filters: []*bloom.BloomFilter{
			bloom.NewWithEstimates(100, 0.01),
			bloom.NewWithEstimates(100, 0.01),
			bloom.NewWithEstimates(100, 0.01),
		},
	}
	b.ready.Store(true)
	return b
}
