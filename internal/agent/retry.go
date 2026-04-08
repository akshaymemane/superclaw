package agent

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

const (
	retryMaxAttempts = 4
	retryBaseDelay   = 500 * time.Millisecond
	retryMaxDelay    = 30 * time.Second
)

// retryLLM wraps an LLMClient and retries transient failures with
// exponential backoff and full jitter.
// Retryable: 429 rate-limit, 500/502/503/504 server errors, connection resets.
// Non-retryable: context cancellation, 400/401/403/404 client errors.
type retryLLM struct {
	inner LLMClient
}

// withRetry wraps inner with automatic retry logic.
func withRetry(inner LLMClient) LLMClient {
	return &retryLLM{inner: inner}
}

// statusCoder is satisfied by the Anthropic SDK error type.
type statusCoder interface {
	StatusCode() int
}

func (r *retryLLM) Create(ctx context.Context, params anthropic.MessageNewParams) (anthropic.Message, error) {
	var lastErr error
	for attempt := 0; attempt < retryMaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return anthropic.Message{}, ctx.Err()
		}

		msg, err := r.inner.Create(ctx, params)
		if err == nil {
			return msg, nil
		}
		lastErr = err

		if !isRetryable(err) {
			return anthropic.Message{}, err
		}

		delay := backoffDelay(attempt)
		select {
		case <-ctx.Done():
			return anthropic.Message{}, ctx.Err()
		case <-time.After(delay):
		}
	}
	return anthropic.Message{}, lastErr
}

// isRetryable returns true for transient errors worth retrying.
func isRetryable(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var sc statusCoder
	if errors.As(err, &sc) {
		code := sc.StatusCode()
		// 429 = rate limit; 500/502/503/504 = server-side transient
		return code == 429 || code == 500 || code == 502 || code == 503 || code == 504
	}
	// Unknown error type — assume retryable (connection reset, DNS, etc.)
	return true
}

// backoffDelay computes exponential backoff with full jitter.
// attempt 0 → ~0–500ms, 1 → ~0–1s, 2 → ~0–2s, 3 → ~0–4s, capped at 30s.
func backoffDelay(attempt int) time.Duration {
	cap := float64(retryMaxDelay)
	base := float64(retryBaseDelay) * float64(int(1)<<attempt)
	if base > cap {
		base = cap
	}
	jitter := rand.Float64() * base //nolint:gosec — non-crypto use
	return time.Duration(jitter)
}
