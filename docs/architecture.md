# Project Architecture

## Template Goal

This repository is a template for building a Go BFF on AWS Lambda using hexagonal architecture.

The core idea is to separate:

- HTTP input and Lambda-specific details;
- application logic;
- input and output contracts;
- external service integrations;
- observability and configuration.

That allows the BFF to grow without mixing routing, use cases, models, and integrations in the same place.

## High-Level View

Request flow:

1. API Gateway receives the HTTP request.
2. Lambda invokes the `bootstrap` binary.
3. [cmd/bff-orchestrator/main.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/cmd/bff-orchestrator/main.go) creates configuration, logger, metrics, tracer, HTTP clients, and use cases.
4. The Lambda handler in [internal/adapters/primary/lambda/handler.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/adapters/primary/lambda/handler.go) detects whether the event comes from API Gateway v1 or v2.
5. The handler normalizes the request and routes it to a use case.
6. The use case lives in `internal/application` and orchestrates calls to one or more output ports.
7. The outbound adapters in `internal/adapters/secondary/http` call downstream services.
8. The use case aggregates the response and returns a common HTTP payload.

## Directory Structure

### `cmd/bff-orchestrator`

Contains the Lambda process entrypoint.

Responsibilities:

- load configuration;
- initialize observability;
- create HTTP clients;
- build use cases;
- inject dependencies into the handler;
- start `lambda.Start(...)`.

Practical rule:

- wiring belongs here;
- business logic does not.

### `internal/domain`

Contains business models and domain errors.

Current files:

- [internal/domain/product.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/domain/product.go)
- [internal/domain/order.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/domain/order.go)
- [internal/domain/errors.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/domain/errors.go)

This layer should contain:

- entities;
- value objects;
- reusable errors;
- aggregated output structures.

This layer should not contain:

- HTTP requests;
- API Gateway structs;
- infrastructure clients.

### `internal/ports`

Defines system contracts.

Current files:

- [internal/ports/inbound.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/ports/inbound.go)
- [internal/ports/outbound.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/ports/outbound.go)

It is split into:

- input ports: what the system can execute;
- output ports: what the system needs from the outside.

Examples:

- `ProductSummaryUseCase` and `OrdersUseCase` are input ports.
- `CatalogService`, `PricingService`, `OrderService`, and `ShipmentService` are output ports.

Practical rule:

- if a use case needs to talk to an external service, model the contract here first.

### `internal/application`

Contains use cases.

Current files:

- [internal/application/get_product_summary.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/application/get_product_summary.go)
- [internal/application/list_orders.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/application/list_orders.go)
- [internal/application/http_response.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/application/http_response.go)

Responsibilities:

- validate inputs;
- coordinate calls across multiple ports;
- apply timeouts;
- aggregate responses;
- transform technical errors into business/API responses.

Current examples:

- `GetProductSummaryUseCase` queries catalog, pricing, and inventory in parallel to build one response.
- `ListOrdersUseCase` retrieves orders and then enriches each one with shipment status in parallel.

Practical rule:

- the use case decides the flow;
- the adapter only performs technical details.

### `internal/adapters/primary`

Contains inbound adapters.

Current file:

- [internal/adapters/primary/lambda/handler.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/adapters/primary/lambda/handler.go)

The current handler:

- supports API Gateway v1 and v2;
- normalizes the request into an `apiRequest`;
- routes using a declarative route table;
- resolves dynamic parameters such as `/products/{productId}`.

Why it matters:

- as the project grows, adding routes does not require duplicated logic across v1 and v2;
- routing stays centralized and maintainable.

### `internal/adapters/secondary`

Contains outbound integrations.

Current files:

- [internal/adapters/secondary/http/product_clients.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/adapters/secondary/http/product_clients.go)
- [internal/adapters/secondary/http/order_clients.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/adapters/secondary/http/order_clients.go)

Responsibilities:

- build HTTP requests;
- call downstream services;
- decode responses;
- return domain models.

Practical rule:

- business orchestration does not belong here;
- use case contracts do not belong here.

### `internal/platform`

Contains cross-cutting capabilities.

Current subdirectories:

- [internal/platform/config/config.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/platform/config/config.go)
- [internal/platform/observability/logger.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/platform/observability/logger.go)
- [internal/platform/observability/metrics.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/platform/observability/metrics.go)
- [internal/platform/observability/tracing.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/platform/observability/tracing.go)
- [internal/platform/observability/http.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/platform/observability/http.go)

Responsibilities:

- read environment variables;
- expose structured logging;
- record metrics;
- record spans;
- instrument outbound HTTP requests.

## Current Use Cases

### `GET /products/{productId}`

Goal:

- build a product summary for the frontend.

Dependencies:

- catalog service;
- pricing service;
- inventory service.

Behavior:

- validates `productId`;
- applies a timeout;
- calls three services in parallel;
- returns `200`, `400`, `502`, or `504` depending on the scenario.

### `GET /orders`

Goal:

- list orders and enrich them with shipment status.

Dependencies:

- order service;
- shipment service.

Behavior:

- lists orders;
- performs controlled fan-out to shipment for each order;
- aggregates the response;
- returns `200`, `502`, or `504`.

## Routing

Routing is defined inside the handler constructor.

Each route contains:

- HTTP method;
- route pattern;
- function that invokes the matching use case.

Conceptual example:

```go
{
    method:  http.MethodGet,
    pattern: "/orders",
    handler: func(ctx context.Context, request apiRequest) (responsePayload, error) {
        return ...
    },
}
```

Benefits:

- adding new routes is incremental;
- duplication between API Gateway v1 and v2 is avoided;
- parameter matching stays centralized.

## Observability

The template already includes three pieces:

- structured logs with `slog`;
- metrics recorded from use cases;
- simple spans to measure execution and relevant calls.

In addition, the HTTP client uses an instrumented `RoundTripper` to record:

- method;
- URL;
- status code;
- duration;
- errors.

Practical rule:

- every new use case should record at least one error counter and one duration metric.

## Configuration

Configuration is read from environment variables and centralized in `config.Load()`.

Current variables:

- `APP_NAME`
- `APP_ENV`
- `LOG_LEVEL`
- `DOWNSTREAM_TIMEOUT_MS`
- `SERVICE_CATALOG_BASE_URL`
- `SERVICE_PRICING_BASE_URL`
- `SERVICE_INVENTORY_BASE_URL`
- `SERVICE_ORDER_BASE_URL`
- `SERVICE_SHIPMENT_BASE_URL`

Practical rule:

- if you add a new downstream, add its base URL here and propagate it through `serverless.yml`.

## Deployment

Deployment is defined in [serverless.yml](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/serverless.yml).

Key points:

- uses Serverless Framework `3.33.0`;
- deploys one Lambda with `provided.al2023`;
- exposes two endpoints through HTTP API;
- injects environment variables;
- enables tracing.

## Recommended Conventions

- use use case names with business verbs: `Get...`, `List...`, `Create...`.
- use capability-oriented output port names: `CatalogService`, `ShipmentService`.
- model aggregated responses in `internal/domain`.
- keep wiring in `cmd/bff-orchestrator/main.go`.
- add tests in `internal/application`.

## Where to Make Each Change

- new endpoint: handler + `serverless.yml`
- new use case: `internal/application`
- new model: `internal/domain`
- new contract: `internal/ports`
- new client: `internal/adapters/secondary`
- new environment variable: `internal/platform/config` + `serverless.yml`
- new wiring: `cmd/bff-orchestrator/main.go`
