package application

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/erickGoJi/hexagonal-aws-template/internal/domain"
	"github.com/erickGoJi/hexagonal-aws-template/internal/platform/observability"
	"github.com/erickGoJi/hexagonal-aws-template/internal/ports"
)

type stubOrderService struct {
	result []domain.Order
	err    error
}

func (s stubOrderService) ListOrders(_ context.Context) ([]domain.Order, error) {
	return s.result, s.err
}

type stubShipmentService struct {
	results map[string]domain.Shipment
	err     error
}

func (s stubShipmentService) GetShipment(_ context.Context, orderID string) (domain.Shipment, error) {
	if s.err != nil {
		return domain.Shipment{}, s.err
	}

	return s.results[orderID], nil
}

func TestListOrdersUseCaseExecute(t *testing.T) {
	logger := slog.Default()
	now := time.Now().UTC()
	useCase := NewListOrdersUseCase(
		stubOrderService{
			result: []domain.Order{
				{
					ID:         "o-1",
					CustomerID: "c-1",
					Status:     "confirmed",
					Total:      150.5,
					Currency:   "USD",
					CreatedAt:  now,
				},
			},
		},
		stubShipmentService{
			results: map[string]domain.Shipment{
				"o-1": {
					OrderID: "o-1",
					Status:  "in_transit",
				},
			},
		},
		logger,
		observability.NewMetrics(logger, "test"),
		observability.NewTracer(logger),
		200*time.Millisecond,
	)

	output, err := useCase.Execute(context.Background(), ports.ListOrdersInput{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.StatusCode != 200 {
		t.Fatalf("unexpected status code: %d", output.StatusCode)
	}
}

func TestListOrdersUseCaseExecuteDownstreamError(t *testing.T) {
	logger := slog.Default()
	useCase := NewListOrdersUseCase(
		stubOrderService{
			err: errors.New("order service unavailable"),
		},
		stubShipmentService{},
		logger,
		observability.NewMetrics(logger, "test"),
		observability.NewTracer(logger),
		200*time.Millisecond,
	)

	output, err := useCase.Execute(context.Background(), ports.ListOrdersInput{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.StatusCode != 502 {
		t.Fatalf("unexpected status code: %d", output.StatusCode)
	}
}
