package http

import (
	"context"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"strconv"
	"strings"

	"github.com/erickGoJi/hexagonal-aws-template/internal/domain"
	"github.com/erickGoJi/hexagonal-aws-template/internal/platform/httpclient"
)

type CatalogClient struct {
	baseURL string
	client  *nethttp.Client
	logger  *slog.Logger
}

type PricingClient struct {
	baseURL string
	client  *nethttp.Client
	logger  *slog.Logger
}

type InventoryClient struct {
	baseURL string
	client  *nethttp.Client
	logger  *slog.Logger
}

type dummyJSONProduct struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
}

func NewCatalogClient(baseURL string, client *nethttp.Client, logger *slog.Logger) *CatalogClient {
	return &CatalogClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
		logger:  logger,
	}
}

func NewPricingClient(baseURL string, client *nethttp.Client, logger *slog.Logger) *PricingClient {
	return &PricingClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
		logger:  logger,
	}
}

func NewInventoryClient(baseURL string, client *nethttp.Client, logger *slog.Logger) *InventoryClient {
	return &InventoryClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
		logger:  logger,
	}
}

func (c *CatalogClient) GetProduct(ctx context.Context, productID string) (domain.CatalogProduct, error) {
	payload, err := fetchDummyJSONProduct(ctx, c.client, c.baseURL, productID)
	if err != nil {
		return domain.CatalogProduct{}, fmt.Errorf("catalog service: %w", err)
	}

	return domain.CatalogProduct{
		ID:          strconv.Itoa(payload.ID),
		Name:        payload.Title,
		Description: payload.Description,
		Tags:        payload.Tags,
	}, nil
}

func (c *PricingClient) GetPrice(ctx context.Context, productID string) (domain.ProductPricing, error) {
	payload, err := fetchDummyJSONProduct(ctx, c.client, c.baseURL, productID)
	if err != nil {
		return domain.ProductPricing{}, fmt.Errorf("pricing service: %w", err)
	}

	return domain.ProductPricing{
		ProductID: strconv.Itoa(payload.ID),
		Amount:    payload.Price,
		Currency:  "USD",
	}, nil
}

func (c *InventoryClient) GetInventory(ctx context.Context, productID string) (domain.ProductInventory, error) {
	payload, err := fetchDummyJSONProduct(ctx, c.client, c.baseURL, productID)
	if err != nil {
		return domain.ProductInventory{}, fmt.Errorf("inventory service: %w", err)
	}

	return domain.ProductInventory{
		ProductID: strconv.Itoa(payload.ID),
		Available: payload.Stock > 0,
		Units:     payload.Stock,
	}, nil
}

func fetchDummyJSONProduct(ctx context.Context, client *nethttp.Client, baseURL, productID string) (dummyJSONProduct, error) {
	var payload dummyJSONProduct
	if err := httpclient.GetJSON(ctx, client, baseURL+"/products/"+productID, &payload, httpclient.DefaultRetryConfig); err != nil {
		return dummyJSONProduct{}, err
	}

	return payload, nil
}

