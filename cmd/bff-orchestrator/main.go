package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	lambdahandler "github.com/erickGoJi/hexagonal-aws-template/internal/adapters/primary/lambda"
	httpadapter "github.com/erickGoJi/hexagonal-aws-template/internal/adapters/secondary/http"
	"github.com/erickGoJi/hexagonal-aws-template/internal/application"
	"github.com/erickGoJi/hexagonal-aws-template/internal/platform/config"
	"github.com/erickGoJi/hexagonal-aws-template/internal/platform/observability"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := observability.NewLogger(cfg.LogLevel)
	metrics := observability.NewMetrics(logger, cfg.AppName)
	tracer := observability.NewTracer(logger)

	httpClient := &http.Client{
		Timeout: cfg.DownstreamTimeout,
		Transport: observability.NewRoundTripper(
			http.DefaultTransport,
			logger,
		),
	}

	catalogClient := httpadapter.NewCatalogClient(
		cfg.ServiceCatalogBaseURL,
		httpClient,
		logger,
	)
	pricingClient := httpadapter.NewPricingClient(
		cfg.ServicePricingBaseURL,
		httpClient,
		logger,
	)
	inventoryClient := httpadapter.NewInventoryClient(
		cfg.ServiceInventoryBaseURL,
		httpClient,
		logger,
	)
	orderClient := httpadapter.NewOrderClient(
		cfg.ServiceOrderBaseURL,
		httpClient,
		logger,
	)

	productSummaryUseCase := application.NewGetProductSummaryUseCase(
		catalogClient,
		pricingClient,
		inventoryClient,
		logger,
		metrics,
		tracer,
		cfg.DownstreamTimeout,
	)
	listOrdersUseCase := application.NewListOrdersUseCase(
		orderClient,
		logger,
		metrics,
		tracer,
		cfg.DownstreamTimeout,
	)

	handler := lambdahandler.NewHandler(productSummaryUseCase, listOrdersUseCase, logger)

	lambda.Start(func(ctx context.Context, event json.RawMessage) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		return handler.Handle(ctx, event)
	})
}
