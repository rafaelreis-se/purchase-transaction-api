# Purchase Transaction API - Clean Makefile for Interview
.PHONY: help build run test lint format clean docker docker-build docker-run api-test health dev info

# Default target
help: ## Show available commands
	@echo 'Purchase Transaction API - Available Commands:'
	@echo ''
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# === Core Development ===
build: ## Build the application
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go

run: ## Run the application locally
	@echo "Starting application..."
	go run cmd/server/main.go

test: ## Run all tests with gotestsum
	@echo "Running tests..."
	gotestsum --format testname ./...

lint: ## Run code quality tools
	@echo "🔍 Running linter..."
	golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/ coverage.out coverage.html transactions.db
	go clean

# === Docker ===
docker-build: ## Build production Docker image
	@echo "Building Docker image..."
	docker build -t purchase-transaction-api:latest .

docker-run: docker-build ## Run application in Docker container
	@echo "Running in Docker with production config..."
	docker run --rm -p 8080:8080 --env-file .env.production purchase-transaction-api:latest

docker-dev: docker-build ## Run application in Docker with local .env
	@echo "Running in Docker with local dev config..."
	docker run --rm -p 8080:8080 --env-file .env purchase-transaction-api:latest

docker: docker-build ## Quick Docker build (for CI/deployment prep)

# === API Testing ===
health: ## Check application health
	@echo "Checking health..."
	@curl -f -s http://localhost:8080/health | jq . || echo " Application not running"

api-test: ## Test complete API workflow
	@echo "Testing API workflow..."
	@echo "\n=== Creating Transaction ==="
	@RESPONSE=$$(curl -s -X POST http://localhost:8080/api/v1/transactions \
		-H "Content-Type: application/json" \
		-d '{"description":"Test purchase","date":"2024-01-15T10:30:00Z","amount":100.50}'); \
	echo "$$RESPONSE" | jq '.'; \
	TRANSACTION_ID=$$(echo "$$RESPONSE" | jq -r '.id'); \
	echo "\n=== Getting Transaction Details ==="; \
	curl -s http://localhost:8080/api/v1/transactions/$$TRANSACTION_ID | jq '.'; \
	echo "\n=== Currency Conversions ==="; \
	echo "\nConverting to EUR..."; \
	curl -s -X POST http://localhost:8080/api/v1/transactions/$$TRANSACTION_ID/convert \
		-H "Content-Type: application/json" \
		-d '{"target_currency":"EUR"}' | jq '.'; \
	echo "\nConverting to BRL..."; \
	curl -s -X POST http://localhost:8080/api/v1/transactions/$$TRANSACTION_ID/convert \
		-H "Content-Type: application/json" \
		-d '{"target_currency":"BRL"}' | jq '.'; \
	echo "\nConverting to CAD..."; \
	curl -s -X POST http://localhost:8080/api/v1/transactions/$$TRANSACTION_ID/convert \
		-H "Content-Type: application/json" \
		-d '{"target_currency":"CAD"}' | jq '.'; \
	echo "\n---OK--- | API workflow complete!"

# === Quick Workflows ===
dev: clean test build ## Quick development cycle
	@echo "OK - Development cycle complete!"

install-tools: ## Install required development tools
	@echo "Installing tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install gotest.tools/gotestsum@latest

info: ## Show project information
	@echo "Purchase Transaction API"
	@echo "======================="
	@echo "Quick Start: make run"
	@echo "Test API:    make api-test"
	@echo "Health:     make health"
	@echo "Build:       make build"