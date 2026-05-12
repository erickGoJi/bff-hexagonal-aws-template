package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	nethttp "net/http"
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
	var payload struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := doRequest(ctx, c.client, c.baseURL+"/products/"+productID, &payload); err != nil {
		return domain.CatalogProduct{}, fmt.Errorf("catalog service: %w", err)
	}

	return domain.CatalogProduct{
		ID:          payload.ID,
		Name:        payload.Name,
		Description: payload.Description,
		Tags:        payload.Tags,
	}, nil
}

func (c *PricingClient) GetPrice(ctx context.Context, productID string) (domain.ProductPricing, error) {
	var payload struct {
		ProductID string  `json:"productId"`
		Amount    float64 `json:"amount"`
		Currency  string  `json:"currency"`
	}

	if err := doRequest(ctx, c.client, c.baseURL+"/prices/"+productID, &payload); err != nil {
		return domain.ProductPricing{}, fmt.Errorf("pricing service: %w", err)
	}

	return domain.ProductPricing{
		ProductID: payload.ProductID,
		Amount:    payload.Amount,
		Currency:  payload.Currency,
	}, nil
}

func (c *InventoryClient) GetInventory(ctx context.Context, productID string) (domain.ProductInventory, error) {
	var payload struct {
		ProductID string `json:"productId"`
		Available bool   `json:"available"`
		Units     int    `json:"units"`
	}

	if err := doRequest(ctx, c.client, c.baseURL+"/inventory/"+productID, &payload); err != nil {
		return domain.ProductInventory{}, fmt.Errorf("inventory service: %w", err)
	}

	return domain.ProductInventory{
		ProductID: payload.ProductID,
		Available: payload.Available,
		Units:     payload.Units,
	}, nil
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
