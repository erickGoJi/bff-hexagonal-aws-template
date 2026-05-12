package observability

import (
	"context"
	"log/slog"
	"time"
)

type Tracer struct {
	logger *slog.Logger
}

type Span struct {
	name      string
	logger    *slog.Logger
	startTime time.Time
}

func NewTracer(logger *slog.Logger) Tracer {
	return Tracer{logger: logger}
}

func (t Tracer) Start(ctx context.Context, name string, attrs ...any) (context.Context, Span) {
	fields := []any{"span", name}
	fields = append(fields, attrs...)
	t.logger.InfoContext(ctx, "span started", fields...)

	return ctx, Span{
		name:      name,
		logger:    t.logger,
		startTime: time.Now(),
	}
}

func (s Span) End(attrs ...any) {
	fields := []any{
		"span", s.name,
		"duration_ms", time.Since(s.startTime).Milliseconds(),
	}
	fields = append(fields, attrs...)
	s.logger.Info("span ended", fields...)
}
