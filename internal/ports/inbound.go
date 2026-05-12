package ports

import "context"

type GetProductSummaryInput struct {
	ProductID string
}

type GetProductSummaryOutput struct {
	StatusCode int
	Body       any
	Headers    map[string]string
}

type ProductSummaryUseCase interface {
	Execute(ctx context.Context, input GetProductSummaryInput) (GetProductSummaryOutput, error)
}

type ListOrdersInput struct{}

type ListOrdersOutput struct {
	StatusCode int
	Body       any
	Headers    map[string]string
}

type OrdersUseCase interface {
	Execute(ctx context.Context, input ListOrdersInput) (ListOrdersOutput, error)
}
