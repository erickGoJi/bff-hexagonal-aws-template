package http

import (
	"context"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/erickGoJi/hexagonal-aws-template/internal/domain"
)

type OrderClient struct {
	baseURL string
	client  *nethttp.Client
	logger  *slog.Logger
}

type ShipmentClient struct {
	baseURL string
	client  *nethttp.Client
	logger  *slog.Logger
}

func NewOrderClient(baseURL string, client *nethttp.Client, logger *slog.Logger) *OrderClient {
	return &OrderClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
		logger:  logger,
	}
}

func NewShipmentClient(baseURL string, client *nethttp.Client, logger *slog.Logger) *ShipmentClient {
	return &ShipmentClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
		logger:  logger,
	}
}

func (c *OrderClient) ListOrders(ctx context.Context) ([]domain.Order, error) {
	var payload []struct {
		ID         string    `json:"id"`
		CustomerID string    `json:"customerId"`
		Status     string    `json:"status"`
		Total      float64   `json:"total"`
		Currency   string    `json:"currency"`
		CreatedAt  time.Time `json:"createdAt"`
	}

	if err := doRequest(ctx, c.client, c.baseURL+"/orders", &payload); err != nil {
		return nil, fmt.Errorf("order service: %w", err)
	}

	orders := make([]domain.Order, 0, len(payload))
	for _, item := range payload {
		orders = append(orders, domain.Order{
			ID:         item.ID,
			CustomerID: item.CustomerID,
			Status:     item.Status,
			Total:      item.Total,
			Currency:   item.Currency,
			CreatedAt:  item.CreatedAt,
		})
	}

	return orders, nil
}

func (c *ShipmentClient) GetShipment(ctx context.Context, orderID string) (domain.Shipment, error) {
	var payload struct {
		OrderID string `json:"orderId"`
		Status  string `json:"status"`
	}

	if err := doRequest(ctx, c.client, c.baseURL+"/shipments/"+orderID, &payload); err != nil {
		return domain.Shipment{}, fmt.Errorf("shipment service: %w", err)
	}

	return domain.Shipment{
		OrderID: payload.OrderID,
		Status:  payload.Status,
	}, nil
}
