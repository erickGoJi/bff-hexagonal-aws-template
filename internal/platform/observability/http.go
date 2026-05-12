package observability

import (
	"log/slog"
	"net/http"
	"time"
)

type loggingRoundTripper struct {
	next   http.RoundTripper
	logger *slog.Logger
}

func NewRoundTripper(next http.RoundTripper, logger *slog.Logger) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}

	return loggingRoundTripper{
		next:   next,
		logger: logger,
	}
}

func (l loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := l.next.RoundTrip(req)
	if err != nil {
		l.logger.ErrorContext(req.Context(), "downstream request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
		)
		return nil, err
	}

	l.logger.InfoContext(req.Context(), "downstream request completed",
		"method", req.Method,
		"url", req.URL.String(),
		"status_code", resp.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return resp, nil
}
