package domain

type ProductSummary struct {
	ProductID   string   `json:"productId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       Price    `json:"price"`
	Stock       Stock    `json:"stock"`
	Tags        []string `json:"tags,omitempty"`
}

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Stock struct {
	Available bool `json:"available"`
	Units     int  `json:"units"`
}

type CatalogProduct struct {
	ID          string
	Name        string
	Description string
	Tags        []string
}

type ProductPricing struct {
	ProductID string
	Amount    float64
	Currency  string
}

type ProductInventory struct {
	ProductID string
	Available bool
	Units     int
}
