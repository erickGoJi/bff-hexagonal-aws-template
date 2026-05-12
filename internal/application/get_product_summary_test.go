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

type stubCatalogService struct {
	result domain.CatalogProduct
	err    error
}

func (s stubCatalogService) GetProduct(_ context.Context, _ string) (domain.CatalogProduct, error) {
	return s.result, s.err
}

type stubPricingService struct {
	result domain.ProductPricing
	err    error
}

func (s stubPricingService) GetPrice(_ context.Context, _ string) (domain.ProductPricing, error) {
	return s.result, s.err
}

type stubInventoryService struct {
	result domain.ProductInventory
	err    error
}

func (s stubInventoryService) GetInventory(_ context.Context, _ string) (domain.ProductInventory, error) {
	return s.result, s.err
}

func TestGetProductSummaryUseCaseExecute(t *testing.T) {
	logger := slog.Default()
	useCase := NewGetProductSummaryUseCase(
		stubCatalogService{
			result: domain.CatalogProduct{
				ID:          "p-1",
				Name:        "Keyboard",
				Description: "Mechanical keyboard",
				Tags:        []string{"peripherals"},
			},
		},
		stubPricingService{
			result: domain.ProductPricing{
				ProductID: "p-1",
				Amount:    99.90,
				Currency:  "USD",
			},
		},
		stubInventoryService{
			result: domain.ProductInventory{
				ProductID: "p-1",
				Available: true,
				Units:     12,
			},
		},
		logger,
		observability.NewMetrics(logger, "test"),
		observability.NewTracer(logger),
		200*time.Millisecond,
	)

	output, err := useCase.Execute(context.Background(), ports.GetProductSummaryInput{ProductID: "p-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.StatusCode != 200 {
		t.Fatalf("unexpected status code: %d", output.StatusCode)
	}
}

func TestGetProductSummaryUseCaseExecuteDownstreamError(t *testing.T) {
	logger := slog.Default()
	useCase := NewGetProductSummaryUseCase(
		stubCatalogService{
			err: errors.New("catalog unavailable"),
		},
		stubPricingService{},
		stubInventoryService{},
		logger,
		observability.NewMetrics(logger, "test"),
		observability.NewTracer(logger),
		200*time.Millisecond,
	)

	output, err := useCase.Execute(context.Background(), ports.GetProductSummaryInput{ProductID: "p-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.StatusCode != 502 {
		t.Fatalf("unexpected status code: %d", output.StatusCode)
	}
}
