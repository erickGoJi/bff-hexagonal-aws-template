package application

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/erickGoJi/hexagonal-aws-template/internal/domain"
	"github.com/erickGoJi/hexagonal-aws-template/internal/platform/observability"
	"github.com/erickGoJi/hexagonal-aws-template/internal/ports"
)

type ListOrdersUseCase struct {
	orderClient      ports.OrderService
	shipmentClient   ports.ShipmentService
	logger           *slog.Logger
	metrics          observability.Metrics
	tracer           observability.Tracer
	timeout          time.Duration
	maxParallelFetch int
}

func NewListOrdersUseCase(
	orderClient ports.OrderService,
	shipmentClient ports.ShipmentService,
	logger *slog.Logger,
	metrics observability.Metrics,
	tracer observability.Tracer,
	timeout time.Duration,
) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		orderClient:      orderClient,
		shipmentClient:   shipmentClient,
		logger:           logger,
		metrics:          metrics,
		tracer:           tracer,
		timeout:          timeout,
		maxParallelFetch: 8,
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

	summaries := make([]domain.OrderSummary, len(orders))
	errs := make(chan error, len(orders))
	sem := make(chan struct{}, uc.maxParallelFetch)
	var wg sync.WaitGroup

	for idx, order := range orders {
		wg.Add(1)

		go func(i int, current domain.Order) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			}
			defer func() { <-sem }()

			shipment, shipmentErr := uc.shipmentClient.GetShipment(ctx, current.ID)
			if shipmentErr != nil {
				errs <- shipmentErr
				return
			}

			summaries[i] = domain.OrderSummary{
				OrderID:        current.ID,
				CustomerID:     current.CustomerID,
				Status:         current.Status,
				ShipmentStatus: shipment.Status,
				Total:          current.Total,
				Currency:       current.Currency,
				CreatedAt:      current.CreatedAt,
			}
		}(idx, order)
	}

	wg.Wait()
	close(errs)

	for fetchErr := range errs {
		if fetchErr == nil {
			continue
		}

		if errors.Is(fetchErr, context.DeadlineExceeded) || errors.Is(fetchErr, context.Canceled) {
			uc.metrics.IncrementCounter(ctx, "orders.list.timeout")
			return ports.ListOrdersOutput{
				StatusCode: http.StatusGatewayTimeout,
				Body: map[string]string{
					"message": "request timed out",
				},
				Headers: defaultHeaders(),
			}, nil
		}

		uc.logger.ErrorContext(ctx, "failed to enrich orders", "error", fetchErr)
		uc.metrics.IncrementCounter(ctx, "orders.list.downstream_error", "source", "shipments")

		return ports.ListOrdersOutput{
			StatusCode: http.StatusBadGateway,
			Body: map[string]string{
				"message": domain.ErrDownstream.Error(),
			},
			Headers: defaultHeaders(),
		}, nil
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
