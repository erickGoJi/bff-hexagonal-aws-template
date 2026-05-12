# Hexagonal AWS Lambda BFF Template

Base template for building a Go BFF on AWS Lambda using the hexagonal architecture pattern.

## Purpose

This project is designed for endpoints that:

- receive a request from the frontend;
- call multiple downstream services;
- aggregate and transform the response;
- expose observability from day one.

## Base Endpoints

- `GET /products/{productId}` aggregates catalog, pricing, and inventory.
- `GET /orders` lists orders and enriches each one with shipment status.

## Structure

```text
cmd/bff-orchestrator/           Lambda entrypoint
internal/application/           Use cases
internal/domain/                Business models and rules
internal/ports/                 Input and output ports
internal/adapters/primary/      Inbound adapters
internal/adapters/secondary/    Outbound adapters
internal/platform/              Configuration and observability
```

## Base Flow

1. API Gateway receives the HTTP request.
2. The Lambda handler translates the request into a use case input.
3. The use case calls N services in parallel when needed.
4. The aggregated response is returned to the client with logs, metrics, and spans.

## Environment Variables

```bash
APP_NAME=hexagonal-bff
APP_ENV=local
LOG_LEVEL=INFO
DOWNSTREAM_TIMEOUT_MS=1500
SERVICE_CATALOG_BASE_URL=https://dummyjson.com
SERVICE_PRICING_BASE_URL=https://dummyjson.com
SERVICE_INVENTORY_BASE_URL=https://dummyjson.com
SERVICE_ORDER_BASE_URL=https://dummyjson.com
```

## Run Tests

```bash
make test
```

## Deployment with Serverless Framework

The project includes [serverless.yml](./serverless.yml) to deploy a Lambda behind HTTP API with these endpoints:

- `GET /products/{productId}`
- `GET /orders`

It uses Serverless Framework `3.33.0`.

Available commands in [Makefile](./Makefile):

```bash
make build
make test
make package
make deploy
```

You can also parameterize stage and region:

```bash
make package STAGE=dev AWS_REGION=us-east-1
make deploy STAGE=dev AWS_REGION=us-east-1
```

## Additional Documentation

Detailed project documentation lives in [docs/README.md](./docs/README.md).
# bff-hexagonal-aws-template
