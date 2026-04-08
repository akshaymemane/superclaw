package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	fetchMaxBodyBytes  = 32 * 1024 // 32 KB
	fetchTimeout       = 15 * time.Second
	fetchRetryAttempts = 3
	fetchRetryBase     = 300 * time.Millisecond
)

// FetchTool performs an HTTP GET request and returns the response body.
// MaxCalls caps the total number of fetches per run; 0 means unlimited.
type FetchTool struct {
	MaxCalls int
	mu       sync.Mutex
	calls    int
}

func NewFetchTool() *FetchTool { return &FetchTool{} }

// WithMaxCalls returns a new FetchTool with the given per-run call limit.
func NewFetchToolWithLimit(maxCalls int) *FetchTool {
	return &FetchTool{MaxCalls: maxCalls}
}

func (f *FetchTool) Name() string { return "fetch_url" }
func (f *FetchTool) Description() string {
	return "Fetch the content of an HTTP or HTTPS URL. Returns up to 32 KB of the response body."
}
func (f *FetchTool) Properties() map[string]any {
	return map[string]any{
		"url": map[string]any{
			"type":        "string",
			"description": "The HTTP or HTTPS URL to fetch.",
		},
	}
}

type fetchInput struct {
	URL string `json:"url"`
}

func (f *FetchTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in fetchInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if in.URL == "" {
		return "", fmt.Errorf("url is required")
	}

	if f.MaxCalls > 0 {
		f.mu.Lock()
		f.calls++
		current := f.calls
		f.mu.Unlock()
		if current > f.MaxCalls {
			return "", fmt.Errorf("fetch_url limit reached (%d calls)", f.MaxCalls)
		}
	}

	u, err := url.Parse(in.URL)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("only http and https URLs are supported, got %q", u.Scheme)
	}

	return fetchWithRetry(ctx, in.URL)
}

func fetchWithRetry(ctx context.Context, rawURL string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < fetchRetryAttempts; attempt++ {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		result, statusCode, err := doFetch(ctx, rawURL)
		if err == nil {
			return result, nil
		}
		lastErr = err
		// Retry only on transient server errors, not client errors or context cancels.
		if !isFetchRetryable(err, statusCode) {
			return "", err
		}
		delay := fetchBackoff(attempt)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(delay):
		}
	}
	return "", lastErr
}

func doFetch(ctx context.Context, rawURL string) (result string, statusCode int, err error) {
	tctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(tctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "superclaw/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, fetchMaxBodyBytes+1))
	if err != nil {
		return "", resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	body := strings.TrimSpace(string(data))
	truncated := ""
	if len(data) > fetchMaxBodyBytes {
		body = strings.TrimSpace(string(data[:fetchMaxBodyBytes]))
		truncated = "\n[truncated]"
	}
	if resp.StatusCode >= 500 {
		return "", resp.StatusCode, fmt.Errorf("server error HTTP %d", resp.StatusCode)
	}
	return fmt.Sprintf("HTTP %d\n%s%s", resp.StatusCode, body, truncated), resp.StatusCode, nil
}

func isFetchRetryable(err error, statusCode int) bool {
	if err == nil {
		return false
	}
	if statusCode >= 500 {
		return true
	}
	// Network-level errors (no status code) are retryable.
	return statusCode == 0
}

func fetchBackoff(attempt int) time.Duration {
	base := float64(fetchRetryBase) * float64(int(1)<<attempt)
	if max := float64(10 * time.Second); base > max {
		base = max
	}
	return time.Duration(rand.Float64() * base) //nolint:gosec
}
