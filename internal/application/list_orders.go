package application

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/erickGoJi/hexagonal-aws-template/internal/domain"
	"github.com/erickGoJi/hexagonal-aws-template/internal/platform/observability"
	"github.com/erickGoJi/hexagonal-aws-template/internal/ports"
)

type ListOrdersUseCase struct {
	orderClient ports.OrderService
	logger      *slog.Logger
	metrics     observability.Metrics
	tracer      observability.Tracer
	timeout     time.Duration
}

func NewListOrdersUseCase(
	orderClient ports.OrderService,
	logger *slog.Logger,
	metrics observability.Metrics,
	tracer observability.Tracer,
	timeout time.Duration,
) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		orderClient: orderClient,
		logger:      logger,
		metrics:     metrics,
		tracer:      tracer,
		timeout:     timeout,
	}
}

func (uc *ListOrdersUseCase) Execute(
	ctx context.Context,
	_ ports.ListOrdersInput,
) (ports.ListOrdersOutput, error) {
	start := time.Now()
	ctx, span := uc.tracer.Start(ctx, "ListOrdersUseCase.Execute")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	orders, err := uc.orderClient.ListOrders(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			uc.metrics.IncrementCounter(ctx, "orders.list.timeout")
			return ports.ListOrdersOutput{
				StatusCode: http.StatusGatewayTimeout,
				Body: map[string]string{
					"message": "request timed out",
				},
				Headers: defaultHeaders(),
			}, nil
		}

		uc.logger.ErrorContext(ctx, "failed to list orders", "error", err)
		uc.metrics.IncrementCounter(ctx, "orders.list.downstream_error", "source", "orders")

		return ports.ListOrdersOutput{
			StatusCode: http.StatusBadGateway,
			Body: map[string]string{
				"message": domain.ErrDownstream.Error(),
			},
			Headers: defaultHeaders(),
		}, nil
	}

	summaries := make([]domain.OrderSummary, 0, len(orders))
	for _, current := range orders {
		summaries = append(summaries, domain.OrderSummary{
			OrderID:        current.ID,
			CustomerID:     current.CustomerID,
			Status:         current.Status,
			ShipmentStatus: current.ShipmentStatus,
			Total:          current.Total,
			Currency:       current.Currency,
			CreatedAt:      current.CreatedAt,
		})
	}

	uc.metrics.RecordDuration(ctx, "orders.list.duration", time.Since(start), "orders_count", len(summaries))

	return ports.ListOrdersOutput{
		StatusCode: http.StatusOK,
		Body: map[string]any{
			"orders": summaries,
			"count":  len(summaries),
		},
		Headers: defaultHeaders(),
	}, nil
}
