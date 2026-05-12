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

func TestListOrdersUseCaseExecute(t *testing.T) {
	logger := slog.Default()
	now := time.Now().UTC()
	useCase := NewListOrdersUseCase(
		stubOrderService{
			result: []domain.Order{
				{
					ID:             "o-1",
					CustomerID:     "c-1",
					Status:         "open",
					ShipmentStatus: "processing",
					Total:          150.5,
					Currency:       "USD",
					CreatedAt:      now,
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
