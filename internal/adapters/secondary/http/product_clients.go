package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	nethttp "net/http"
	"strconv"
	"strings"

	"github.com/erickGoJi/hexagonal-aws-template/internal/domain"
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
	if err := doRequest(ctx, client, baseURL+"/products/"+productID, &payload); err != nil {
		return dummyJSONProduct{}, err
	}

	return payload, nil
}

func doRequest(ctx context.Context, client *nethttp.Client, url string, target any) error {
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= nethttp.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
