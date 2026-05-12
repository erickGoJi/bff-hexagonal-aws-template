package observability

import (
	"context"
	"log/slog"
	"time"
)

type Metrics struct {
	logger      *slog.Logger
	serviceName string
}

func NewMetrics(logger *slog.Logger, serviceName string) Metrics {
	return Metrics{
		logger:      logger,
		serviceName: serviceName,
	}
}

func (m Metrics) RecordDuration(_ context.Context, name string, duration time.Duration, attrs ...any) {
	fields := []any{
		"metric_type", "duration",
		"metric_name", name,
		"duration_ms", duration.Milliseconds(),
		"service", m.serviceName,
	}
	fields = append(fields, attrs...)
	m.logger.Info("metric recorded", fields...)
}

func (m Metrics) IncrementCounter(_ context.Context, name string, attrs ...any) {
	fields := []any{
		"metric_type", "counter",
		"metric_name", name,
		"value", 1,
		"service", m.serviceName,
	}
	fields = append(fields, attrs...)
	m.logger.Info("metric recorded", fields...)
}
