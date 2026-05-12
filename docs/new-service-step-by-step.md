# How to Create a New Service Step by Step

## Goal of This Guide

This guide explains how to add a new service or capability to the BFF without breaking the hexagonal structure.

In this context, a "new service" usually means one of these two things:

- a new BFF endpoint;
- a new downstream integration required by a use case.

Most of the time, you will add both at the same time.

## Example We Will Follow

Assume you want to add:

- endpoint: `GET /customers/{customerId}`
- primary downstream: `customer service`
- additional downstream: `loyalty service`

The final goal is to return an aggregated response for the frontend.

## Correct Order Summary

1. Define the use case.
2. Model the domain and contracts.
3. Implement the use case.
4. Implement downstream adapters.
5. Add configuration.
6. Wire everything in `main.go`.
7. Register the route in the handler.
8. Expose the endpoint in `serverless.yml`.
9. Add tests.
10. Update documentation.

## Step 1. Define the Use Case Before Writing Code

Before touching files, make these points explicit:

- which endpoint you will expose;
- which data the frontend needs;
- which downstream services participate;
- whether calls should be sequential or parallel;
- which responses are expected;
- which errors should map to `400`, `404`, `502`, `504`, and so on.

Example:

- endpoint: `GET /customers/{customerId}`
- downstreams: customer and loyalty
- response: customer profile + loyalty tier

## Step 2. Create or Update Domain Models

Add the models in `internal/domain`.

Example:

```go
package domain

type CustomerProfile struct {
    CustomerID string `json:"customerId"`
    Name       string `json:"name"`
    Email      string `json:"email"`
    Tier       string `json:"tier"`
}

type Customer struct {
    ID    string
    Name  string
    Email string
}

type LoyaltyStatus struct {
    CustomerID string
    Tier       string
}
```

Use different models when it makes sense:

- raw downstream model;
- final aggregated model for the frontend.

## Step 3. Define Input and Output Ports

Edit [internal/ports/inbound.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/ports/inbound.go) and [internal/ports/outbound.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/ports/outbound.go).

### Input Port

Example:

```go
type GetCustomerProfileInput struct {
    CustomerID string
}

type GetCustomerProfileOutput struct {
    StatusCode int
    Body       any
    Headers    map[string]string
}

type CustomerProfileUseCase interface {
    Execute(ctx context.Context, input GetCustomerProfileInput) (GetCustomerProfileOutput, error)
}
```

### Output Ports

Example:

```go
type CustomerService interface {
    GetCustomer(ctx context.Context, customerID string) (domain.Customer, error)
}

type LoyaltyService interface {
    GetStatus(ctx context.Context, customerID string) (domain.LoyaltyStatus, error)
}
```

Rule:

- contract first;
- implementation second.

## Step 4. Implement the Use Case

Create a file in `internal/application`, for example `get_customer_profile.go`.

The use case should:

- receive interfaces, not concrete implementations;
- apply a timeout;
- validate input;
- coordinate calls;
- record logs, metrics, and spans;
- map technical errors to coherent HTTP responses.

Suggested structure:

```go
type GetCustomerProfileUseCase struct {
    customerClient ports.CustomerService
    loyaltyClient  ports.LoyaltyService
    logger         *slog.Logger
    metrics        observability.Metrics
    tracer         observability.Tracer
    timeout        time.Duration
}
```

If you need fan-out:

- use goroutines;
- control concurrency if the list can grow;
- consider timeouts and cancellation through context.

## Step 5. Implement Outbound Adapters

Add the clients in `internal/adapters/secondary/http`.

Examples:

- `customer_clients.go`
- `loyalty_clients.go`

Each adapter should:

- build the URL;
- execute the request;
- validate the status code;
- decode JSON;
- return domain models.

Do not put these concerns here:

- business decisions;
- response aggregation;
- BFF HTTP routing.

## Step 6. Add Configuration

Edit [internal/platform/config/config.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/platform/config/config.go).

Add new fields:

```go
ServiceCustomerBaseURL string
ServiceLoyaltyBaseURL  string
```

And load them from environment variables:

```go
ServiceCustomerBaseURL: getEnv("SERVICE_CUSTOMER_BASE_URL", "http://localhost:8086"),
ServiceLoyaltyBaseURL:  getEnv("SERVICE_LOYALTY_BASE_URL", "http://localhost:8087"),
```

Rule:

- every new downstream should have explicit configuration.

## Step 7. Wire Everything in `cmd/bff-orchestrator/main.go`

This is where everything gets connected.

You should:

1. create the downstream HTTP client;
2. create outbound adapters;
3. instantiate the new use case;
4. inject it into the handler.

Conceptual example:

```go
customerClient := httpadapter.NewCustomerClient(cfg.ServiceCustomerBaseURL, httpClient, logger)
loyaltyClient := httpadapter.NewLoyaltyClient(cfg.ServiceLoyaltyBaseURL, httpClient, logger)

customerProfileUseCase := application.NewGetCustomerProfileUseCase(
    customerClient,
    loyaltyClient,
    logger,
    metrics,
    tracer,
    cfg.DownstreamTimeout,
)
```

## Step 8. Register the Route in the Handler

Edit [internal/adapters/primary/lambda/handler.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/adapters/primary/lambda/handler.go).

### 8.1 Add the dependency to `Handler`

Example:

```go
customerProfileUseCase ports.CustomerProfileUseCase
```

### 8.2 Inject it in `NewHandler`

Update the constructor and struct fields.

### 8.3 Add a new entry to `routes`

Example:

```go
{
    method:  http.MethodGet,
    pattern: "/customers/{customerId}",
    handler: func(ctx context.Context, request apiRequest) (responsePayload, error) {
        output, err := customerProfileUseCase.Execute(ctx, ports.GetCustomerProfileInput{
            CustomerID: request.PathParameters["customerId"],
        })
        if err != nil {
            return responsePayload{}, err
        }

        return responsePayload(output), nil
    },
}
```

Rule:

- add a new entry to the route table;
- do not go back to scattered `if` blocks in the handler.

## Step 9. Expose the Endpoint in `serverless.yml`

Add:

```yaml
- httpApi:
    path: /customers/{customerId}
    method: GET
```

If you added new variables, also add them under `provider.environment`.

## Step 10. Add Tests

The minimum expected coverage is in `internal/application`.

You should test:

- successful flow;
- invalid input when applicable;
- downstream error;
- timeout when the use case can trigger it.

Current pattern:

- create small stubs;
- inject them into the use case;
- assert on `StatusCode`.

Reference files:

- [internal/application/get_product_summary_test.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/application/get_product_summary_test.go)
- [internal/application/list_orders_test.go](/Users/erickeduardogomezjimenez/projects/hexagonal-aws-template/internal/application/list_orders_test.go)

## Step 11. Update Documentation

When you add a new capability, update:

- `README.md` if visible behavior changes;
- `docs/architecture.md` if the structure changes;
- this guide if you introduce a new convention.

## Short Checklist

- new domain model created
- new ports defined
- new use case implemented
- new downstream client implemented
- new config added
- new wiring added in `main.go`
- new route added in the handler
- new endpoint added in `serverless.yml`
- tests added
- docs updated

## Common Mistakes

### Putting Business Logic in the HTTP Adapter

Incorrect:

- the adapter decides how to aggregate data from two services.

Correct:

- the adapter only calls and translates;
- the use case orchestrates.

### Skipping Ports

Incorrect:

- the use case depends on a concrete HTTP client.

Correct:

- the use case depends on an interface in `internal/ports`.

### Adding Routes Through Growing `if` Blocks

Incorrect:

- duplicating routing decisions across API Gateway v1 and v2.

Correct:

- adding a new entry to the `routes` table.

### Forgetting to Propagate Configuration to `serverless.yml`

Incorrect:

- adding a variable in `config.go` but forgetting deployment config.

Correct:

- add it in `config.go`, `serverless.yml`, and documentation.

## Useful Commands

```bash
make test
make build
make package STAGE=dev AWS_REGION=us-east-1
make deploy STAGE=dev AWS_REGION=us-east-1
```
