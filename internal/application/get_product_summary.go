package application

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/erickGoJi/hexagonal-aws-template/internal/domain"
	"github.com/erickGoJi/hexagonal-aws-template/internal/platform/observability"
	"github.com/erickGoJi/hexagonal-aws-template/internal/ports"
)

type GetProductSummaryUseCase struct {
	catalogClient   ports.CatalogService
	pricingClient   ports.PricingService
	inventoryClient ports.InventoryService
	logger          *slog.Logger
	metrics         observability.Metrics
	tracer          observability.Tracer
	timeout         time.Duration
}

func NewGetProductSummaryUseCase(
	catalogClient ports.CatalogService,
	pricingClient ports.PricingService,
	inventoryClient ports.InventoryService,
	logger *slog.Logger,
	metrics observability.Metrics,
	tracer observability.Tracer,
	timeout time.Duration,
) *GetProductSummaryUseCase {
	return &GetProductSummaryUseCase{
		catalogClient:   catalogClient,
		pricingClient:   pricingClient,
		inventoryClient: inventoryClient,
		logger:          logger,
		metrics:         metrics,
		tracer:          tracer,
		timeout:         timeout,
	}
}

func (uc *GetProductSummaryUseCase) Execute(
	ctx context.Context,
	input ports.GetProductSummaryInput,
) (ports.GetProductSummaryOutput, error) {
	start := time.Now()
	ctx, span := uc.tracer.Start(ctx, "GetProductSummaryUseCase.Execute", "product_id", input.ProductID)
	defer span.End()

	productID := strings.TrimSpace(input.ProductID)
	if productID == "" {
		uc.metrics.IncrementCounter(ctx, "product_summary.validation_error")
		return ports.GetProductSummaryOutput{
			StatusCode: http.StatusBadRequest,
			Body: map[string]string{
				"message": domain.ErrInvalidProductID.Error(),
			},
			Headers: defaultHeaders(),
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	var (
		catalog   domain.CatalogProduct
		price     domain.ProductPricing
		inventory domain.ProductInventory
		errs      = make(chan error, 3)
		wg        sync.WaitGroup
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		result, err := uc.catalogClient.GetProduct(ctx, productID)
		if err != nil {
			errs <- err
			return
		}
		catalog = result
	}()

	go func() {
		defer wg.Done()
		result, err := uc.pricingClient.GetPrice(ctx, productID)
		if err != nil {
			errs <- err
			return
		}
		price = result
	}()

	go func() {
		defer wg.Done()
		result, err := uc.inventoryClient.GetInventory(ctx, productID)
		if err != nil {
			errs <- err
			return
		}
		inventory = result
	}()

	wg.Wait()
	close(errs)

	for err := range errs {
		if err == nil {
			continue
		}

		uc.logger.ErrorContext(ctx, "failed to build product summary",
			"product_id", productID,
			"error", err,
		)
		uc.metrics.IncrementCounter(ctx, "product_summary.downstream_error", "product_id", productID)

		return ports.GetProductSummaryOutput{
			StatusCode: http.StatusBadGateway,
			Body: map[string]string{
				"message": domain.ErrDownstream.Error(),
			},
			Headers: defaultHeaders(),
		}, nil
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		uc.metrics.IncrementCounter(ctx, "product_summary.timeout", "product_id", productID)
		return ports.GetProductSummaryOutput{
			StatusCode: http.StatusGatewayTimeout,
			Body: map[string]string{
				"message": "request timed out",
			},
			Headers: defaultHeaders(),
		}, nil
	}

	summary := domain.ProductSummary{
		ProductID:   catalog.ID,
		Name:        catalog.Name,
		Description: catalog.Description,
		Price: domain.Price{
			Amount:   price.Amount,
			Currency: price.Currency,
		},
		Stock: domain.Stock{
			Available: inventory.Available,
			Units:     inventory.Units,
		},
		Tags: catalog.Tags,
	}

	uc.metrics.RecordDuration(ctx, "product_summary.duration", time.Since(start), "product_id", productID)

	return ports.GetProductSummaryOutput{
		StatusCode: http.StatusOK,
		Body:       summary,
		Headers:    defaultHeaders(),
	}, nil
}
