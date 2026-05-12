APP_NAME ?= hexagonal-aws-template
STAGE ?= dev
AWS_REGION ?= us-east-1
GOOS ?= linux
GOARCH ?= arm64
CGO_ENABLED ?= 0
LDFLAGS ?= -s -w
BOOTSTRAP_BINARY ?= bootstrap
SLS ?= npx serverless

.PHONY: build test package deploy clean

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o $(BOOTSTRAP_BINARY) ./cmd/bff-orchestrator

test:
	go test ./...

package: build
	$(SLS) package --stage $(STAGE) --region $(AWS_REGION)

deploy: build
	$(SLS) deploy --stage $(STAGE) --region $(AWS_REGION)

clean:
	rm -f $(BOOTSTRAP_BINARY)
