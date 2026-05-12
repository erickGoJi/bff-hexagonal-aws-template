package http

import (
	"context"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/erickGoJi/hexagonal-aws-template/internal/domain"
)

type OrderClient struct {
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

func (c *OrderClient) ListOrders(ctx context.Context) ([]domain.Order, error) {
	var payload struct {
		Carts []struct {
			ID        int     `json:"id"`
			UserID    int     `json:"userId"`
			Total     float64 `json:"total"`
			TotalQty  int     `json:"totalQuantity"`
			CreatedAt string  `json:"createdAt"`
		} `json:"carts"`
	}

	if err := doRequest(ctx, c.client, c.baseURL+"/carts", &payload); err != nil {
		return nil, fmt.Errorf("order service: %w", err)
	}

	orders := make([]domain.Order, 0, len(payload.Carts))
	for _, item := range payload.Carts {
		createdAt := time.Time{}
		if item.CreatedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, item.CreatedAt); err == nil {
				createdAt = parsed
			}
		}

		orderStatus := "open"
		shipmentStatus := "processing"
		if item.TotalQty == 0 {
			orderStatus = "empty"
			shipmentStatus = "pending"
		}

		orders = append(orders, domain.Order{
			ID:             strconv.Itoa(item.ID),
			CustomerID:     strconv.Itoa(item.UserID),
			Status:         orderStatus,
			ShipmentStatus: shipmentStatus,
			Total:          item.Total,
			Currency:       "USD",
			CreatedAt:      createdAt,
		})
	}

	return orders, nil
}
