# Fire PMAAS - Makefile for development and testing

.PHONY: help test test-unit test-integration test-coverage test-verbose clean build run lint format deps docker-build docker-test

# Default target
help: ## Show this help message
	@echo "Fire PMAAS - Property Management as a Service"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Dependencies
deps: ## Download and install dependencies
	go mod download
	go mod tidy

# Testing targets
test: ## Run all tests
	go test -race -v ./...

test-unit: ## Run only unit tests
	go test -race -v ./pkg/models/... ./pkg/middleware/... ./pkg/testutils/...

test-integration: ## Run only integration tests
	go test -race -v ./tests/integration/... ./pkg/api/...

test-coverage: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-verbose: ## Run tests with verbose output
	go test -race -v -count=1 ./...

test-benchmark: ## Run benchmark tests
	go test -bench=. -benchmem ./...

test-clean: ## Clean test cache
	go clean -testcache

# Database testing
test-db: ## Run database-specific tests
	go test -race -v ./pkg/db/... ./pkg/models/...

test-rbac: ## Run role-based access control tests
	go test -race -v ./tests/integration/rbac_test.go

# Code quality
lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

# Build targets
build: ## Build the application
	go build -o bin/fire-pmaas ./cmd/server

run: ## Run the application
	go run ./cmd/server/main.go

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker targets
docker-build: ## Build Docker image
	docker build -t fire-pmaas:latest .

docker-test: ## Run tests in Docker container
	docker build -f Dockerfile.test -t fire-pmaas-test .
	docker run --rm fire-pmaas-test

docker-run: ## Run application in Docker
	docker-compose up --build

docker-down: ## Stop Docker containers
	docker-compose down

# Development environment
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@echo "Installing dependencies..."
	go mod download
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development environment ready!"

dev-db: ## Start development database
	docker-compose up -d postgres

dev-auth: ## Start development authentication (Keycloak)
	docker-compose up -d keycloak

dev-services: ## Start all development services
	docker-compose up -d

# Migration targets
migrate-up: ## Run database migrations up
	go run cmd/migrate/main.go up

migrate-down: ## Run database migrations down
	go run cmd/migrate/main.go down

migrate-reset: ## Reset database (down then up)
	go run cmd/migrate/main.go reset

# Security targets
security-scan: ## Run security scan
	gosec ./...

dependency-check: ## Check for vulnerable dependencies
	go list -json -m all | nancy sleuth

# Performance targets
profile-cpu: ## Run CPU profiling
	go test -cpuprofile cpu.prof -bench=. ./...

profile-mem: ## Run memory profiling
	go test -memprofile mem.prof -bench=. ./...

# Documentation
docs: ## Generate documentation
	godoc -http=:6060
	@echo "Documentation server started at http://localhost:6060"

# Git hooks
install-hooks: ## Install git hooks
	cp scripts/pre-commit .git/hooks/
	chmod +x .git/hooks/pre-commit

# Environment targets
env-test: ## Set up test environment variables
	export POSTGRES_HOST=localhost
	export POSTGRES_PORT=5432
	export POSTGRES_USER=test_user
	export POSTGRES_PASSWORD=test_pass
	export POSTGRES_DB=test_db
	export KEYCLOAK_ISSUER=http://localhost:8080/realms/test

env-dev: ## Set up development environment variables
	export POSTGRES_HOST=localhost
	export POSTGRES_PORT=5432
	export POSTGRES_USER=pmaas_user
	export POSTGRES_PASSWORD=pmaas_pass
	export POSTGRES_DB=pmaas_dev
	export KEYCLOAK_ISSUER=http://localhost:8080/realms/pmaas

# Quick test commands for CI/CD
ci-test: deps test-clean test-coverage lint vet ## Run all CI tests

# Coverage targets
coverage-badge: ## Generate coverage badge
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//' > coverage.txt

# All targets
all: clean deps lint vet test build ## Run all checks and build
