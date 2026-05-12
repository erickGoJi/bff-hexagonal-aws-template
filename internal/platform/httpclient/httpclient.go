package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"strings"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   100 * time.Millisecond,
	MaxDelay:    1 * time.Second,
}

func GetJSON(ctx context.Context, client *nethttp.Client, url string, target any, retryCfg RetryConfig) error {
	if retryCfg.MaxAttempts <= 0 {
		retryCfg = DefaultRetryConfig
	}
	if retryCfg.BaseDelay <= 0 {
		retryCfg.BaseDelay = DefaultRetryConfig.BaseDelay
	}
	if retryCfg.MaxDelay <= 0 {
		retryCfg.MaxDelay = DefaultRetryConfig.MaxDelay
	}

	var lastErr error
	for attempt := 1; attempt <= retryCfg.MaxAttempts; attempt++ {
		err := doGetJSON(ctx, client, url, target)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt == retryCfg.MaxAttempts || !isRetryable(err) {
			return lastErr
		}

		delay := backoffDelay(attempt, retryCfg.BaseDelay, retryCfg.MaxDelay)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

func doGetJSON(ctx context.Context, client *nethttp.Client, url string, target any) error {
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= nethttp.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("http status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func isRetryable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "request failed") ||
		strings.Contains(msg, "http status 408") ||
		strings.Contains(msg, "http status 429") ||
		strings.Contains(msg, "http status 500") ||
		strings.Contains(msg, "http status 502") ||
		strings.Contains(msg, "http status 503") ||
		strings.Contains(msg, "http status 504")
}

func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	d := base * time.Duration(1<<(attempt-1))
	if d > max {
		return max
	}
	return d
}
