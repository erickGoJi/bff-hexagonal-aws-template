package ports

import (
	"context"

	"github.com/erickGoJi/hexagonal-aws-template/internal/domain"
)

type CatalogService interface {
	GetProduct(ctx context.Context, productID string) (domain.CatalogProduct, error)
}

type PricingService interface {
	GetPrice(ctx context.Context, productID string) (domain.ProductPricing, error)
}

type InventoryService interface {
	GetInventory(ctx context.Context, productID string) (domain.ProductInventory, error)
}

type OrderService interface {
	ListOrders(ctx context.Context) ([]domain.Order, error)
}

type ShipmentService interface {
	GetShipment(ctx context.Context, orderID string) (domain.Shipment, error)
}
