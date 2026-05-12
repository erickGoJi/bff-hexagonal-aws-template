package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/erickGoJi/hexagonal-aws-template/internal/ports"
)

type Handler struct {
	productSummaryUseCase ports.ProductSummaryUseCase
	listOrdersUseCase     ports.OrdersUseCase
	logger                *slog.Logger
	routesByMethod        map[string][]compiledRoute
}

func NewHandler(
	productSummaryUseCase ports.ProductSummaryUseCase,
	listOrdersUseCase ports.OrdersUseCase,
	logger *slog.Logger,
) *Handler {
	routes := []route{
		{
			method:  http.MethodGet,
			pattern: "/orders",
			handler: func(ctx context.Context, request apiRequest) (responsePayload, error) {
				output, err := listOrdersUseCase.Execute(ctx, ports.ListOrdersInput{})
				if err != nil {
					return responsePayload{}, err
				}

				return responsePayload(output), nil
			},
		},
		{
			method:  http.MethodGet,
			pattern: "/products/{productId}",
			handler: func(ctx context.Context, request apiRequest) (responsePayload, error) {
				output, err := productSummaryUseCase.Execute(ctx, ports.GetProductSummaryInput{
					ProductID: request.PathParameters["productId"],
				})
				if err != nil {
					return responsePayload{}, err
				}

				return responsePayload(output), nil
			},
		},
	}

	return &Handler{
		productSummaryUseCase: productSummaryUseCase,
		listOrdersUseCase:     listOrdersUseCase,
		logger:                logger,
		routesByMethod:        compileRoutes(routes),
	}
}

func (h *Handler) Handle(ctx context.Context, event json.RawMessage) (any, error) {
	var envelope struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(event, &envelope); err != nil {
		return nil, fmt.Errorf("decode lambda event envelope: %w", err)
	}

	if envelope.Version == "2.0" {
		var request events.APIGatewayV2HTTPRequest
		if err := json.Unmarshal(event, &request); err != nil {
			return nil, fmt.Errorf("decode API Gateway v2 event: %w", err)
		}

		return h.handleAPIGatewayV2(ctx, request)
	}

	var request events.APIGatewayProxyRequest
	if err := json.Unmarshal(event, &request); err != nil {
		h.logger.ErrorContext(ctx, "unsupported lambda event", "error", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"message":"unsupported event type"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	return h.handleAPIGatewayV1(ctx, request)
}

func (h *Handler) handleAPIGatewayV1(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	output, err := h.route(ctx, apiRequest{
		Method:         request.HTTPMethod,
		Path:           request.Path,
		PathParameters: cloneMap(request.PathParameters),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	body, err := json.Marshal(output.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: output.StatusCode,
		Headers:    output.Headers,
		Body:       string(body),
	}, nil
}

func (h *Handler) handleAPIGatewayV2(
	ctx context.Context,
	request events.APIGatewayV2HTTPRequest,
) (events.APIGatewayV2HTTPResponse, error) {
	output, err := h.route(ctx, apiRequest{
		Method:         request.RequestContext.HTTP.Method,
		Path:           request.RawPath,
		PathParameters: cloneMap(request.PathParameters),
	})
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}

	body, err := json.Marshal(output.Body)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: output.StatusCode,
		Headers:    output.Headers,
		Body:       string(body),
	}, nil
}

func (h *Handler) route(ctx context.Context, request apiRequest) (responsePayload, error) {
	requestPathParts := splitPath(request.Path)
	candidates := h.routesByMethod[request.Method]

	for _, candidate := range candidates {
		matched, pathParams := matchCompiledPath(candidate, requestPathParts)
		if !matched {
			continue
		}

		for key, value := range pathParams {
			request.PathParameters[key] = value
		}

		return candidate.handler(ctx, request)
	}

	h.logger.WarnContext(ctx, "route not found", "method", request.Method, "path", request.Path)

	return responsePayload{
		StatusCode: http.StatusNotFound,
		Body: map[string]string{
			"message": "route not found",
		},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func compileRoutes(routes []route) map[string][]compiledRoute {
	compiledByMethod := make(map[string][]compiledRoute)
	for _, candidate := range routes {
		compiledByMethod[candidate.method] = append(
			compiledByMethod[candidate.method],
			compiledRoute{
				pathParts: splitPath(candidate.pattern),
				handler:   candidate.handler,
			},
		)
	}

	return compiledByMethod
}

func matchCompiledPath(candidate compiledRoute, pathParts []string) (bool, map[string]string) {
	if len(candidate.pathParts) != len(pathParts) {
		return false, nil
	}

	params := make(map[string]string)
	for idx := range candidate.pathParts {
		patternPart := candidate.pathParts[idx]
		pathPart := pathParts[idx]

		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			paramName := strings.TrimSuffix(strings.TrimPrefix(patternPart, "{"), "}")
			params[paramName] = pathPart
			continue
		}

		if patternPart != pathPart {
			return false, nil
		}
	}

	return true, params
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}

	return strings.Split(trimmed, "/")
}

func cloneMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}

	return cloned
}

type responsePayload struct {
	StatusCode int
	Body       any
	Headers    map[string]string
}

type apiRequest struct {
	Method         string
	Path           string
	PathParameters map[string]string
}

type route struct {
	method  string
	pattern string
	handler func(ctx context.Context, request apiRequest) (responsePayload, error)
}

type compiledRoute struct {
	pathParts []string
	handler   func(ctx context.Context, request apiRequest) (responsePayload, error)
}
