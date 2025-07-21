# RHOAI Grafana Dashboards Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=dashboard-manager
BINARY_UNIX=$(BINARY_NAME)_unix

# Helm parameters
HELM_CMD=helm
CHART_NAME=grafana-dashboards
RELEASE_NAME=rhoai-dashboards
NAMESPACE=rhoai-observability

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build clean test coverage lint helm-lint helm-template helm-install helm-upgrade helm-uninstall help

all: test build

## Build the CLI tool
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./cmd/dashboard-manager

## Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) -v ./cmd/dashboard-manager

## Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

## Run tests
test:
	$(GOTEST) -v ./...

## Run tests with coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Run linter
lint:
	golangci-lint run

## Download dependencies
deps:
	$(GOMOD) tidy
	$(GOMOD) download

## Lint Helm chart
helm-lint:
	$(HELM_CMD) lint .

## Template Helm chart
helm-template:
	$(HELM_CMD) template $(RELEASE_NAME) . --namespace $(NAMESPACE)

## Template with production values
helm-template-prod:
	$(HELM_CMD) template $(RELEASE_NAME) . --namespace $(NAMESPACE) -f examples/values-production.yaml

## Install Helm chart
helm-install:
	$(HELM_CMD) install $(RELEASE_NAME) . --namespace $(NAMESPACE) --create-namespace

## Upgrade Helm chart
helm-upgrade:
	$(HELM_CMD) upgrade $(RELEASE_NAME) . --namespace $(NAMESPACE)

## Uninstall Helm chart
helm-uninstall:
	$(HELM_CMD) uninstall $(RELEASE_NAME) --namespace $(NAMESPACE)

## Package Helm chart
helm-package:
	$(HELM_CMD) package .

## Validate dashboard files
validate:
	./$(BINARY_NAME) validate

## List dashboard files
list:
	./$(BINARY_NAME) list

## Run all validation tests
test-all: helm-lint test validate

## Install development dependencies
dev-setup:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Run the test script
test-script:
	./scripts/test-chart.sh

## Build Docker image
docker-build:
	docker build -t rhoai/dashboard-manager:$(VERSION) .

## Run in development mode
dev: build
	./$(BINARY_NAME) --help

## Generate documentation
docs:
	@echo "Generating documentation..."
	@mkdir -p docs/generated
	./$(BINARY_NAME) --help > docs/generated/cli-help.txt

## Install chart with monitoring
monitor-install: helm-install
	@echo "Checking deployment status..."
	kubectl get grafanadashboards -n $(NAMESPACE)
	kubectl get pods -n $(NAMESPACE)

## Show status
status:
	@echo "=== Helm Release Status ==="
	$(HELM_CMD) status $(RELEASE_NAME) -n $(NAMESPACE) || echo "Release not found"
	@echo ""
	@echo "=== Dashboard Resources ==="
	kubectl get grafanadashboards -n $(NAMESPACE) -l app.kubernetes.io/instance=$(RELEASE_NAME) || echo "No dashboards found"
	@echo ""
	@echo "=== Grafana Instances ==="
	kubectl get grafana --all-namespaces || echo "No Grafana instances found"

## Show help
help:
	@echo "RHOAI Grafana Dashboards - Available commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build          Build the CLI tool"
	@echo "  build-linux    Build for Linux"
	@echo "  clean          Clean build artifacts"
	@echo ""
	@echo "Development Commands:"
	@echo "  test           Run tests"
	@echo "  coverage       Run tests with coverage"
	@echo "  lint           Run linter"
	@echo "  dev-setup      Install development dependencies"
	@echo "  validate       Validate dashboard files"
	@echo "  list           List dashboard files"
	@echo ""
	@echo "Helm Commands:"
	@echo "  helm-lint      Lint Helm chart"
	@echo "  helm-template  Template Helm chart"
	@echo "  helm-install   Install Helm chart"
	@echo "  helm-upgrade   Upgrade Helm chart"
	@echo "  helm-uninstall Uninstall Helm chart"
	@echo "  helm-package   Package Helm chart"
	@echo ""
	@echo "Testing Commands:"
	@echo "  test-all       Run all validation tests"
	@echo "  test-script    Run the comprehensive test script"
	@echo ""
	@echo "Utility Commands:"
	@echo "  status         Show deployment status"
	@echo "  monitor-install Install with monitoring"
	@echo "  docs           Generate documentation"
	@echo "  help           Show this help message"
