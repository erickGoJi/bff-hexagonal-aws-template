package domain

import "time"

type OrderSummary struct {
	OrderID        string    `json:"orderId"`
	CustomerID     string    `json:"customerId"`
	Status         string    `json:"status"`
	ShipmentStatus string    `json:"shipmentStatus"`
	Total          float64   `json:"total"`
	Currency       string    `json:"currency"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Order struct {
	ID         string
	CustomerID string
	Status     string
	Total      float64
	Currency   string
	CreatedAt  time.Time
}

type Shipment struct {
	OrderID string
	Status  string
}
