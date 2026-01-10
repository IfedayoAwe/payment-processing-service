ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: help
help:
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Common targets:"
	@echo "  run                    Run Go service locally"
	@echo "  test                   Run tests"
	@echo "  lint                   Lint code"
	@echo "  docker-build           Build Docker image"
	@echo "  dev-up                 Start local dev stack via Docker Compose"
	@echo "  dev-down               Stop dev stack"
	@echo "  dev-restart            Restart dev stack without rebuilding"
	@echo "  sqlc-generate          Generate Go code from SQL queries"
	@echo ""

# ======== DEV ========

run:
	go run main.go

test:
	go test ./... -coverprofile=coverage.out
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total: | awk '{print substr($$3, 1, length($$3)-1)}') && \
	echo "Coverage: $$COVERAGE%"

test-verbose:
	go test ./... -v -coverprofile=coverage.out
	@go tool cover -func=coverage.out

# ======== DOCKER ========

dev-up:
	docker compose up --build

dev-down:
	docker compose down

dev-restart:
	docker compose down
	docker compose up -d

docker-build:
	docker build -t payment-service:latest .

# ======== LINTING ========

lint:
	golangci-lint run ./...

# ======== SQLC ========

sqlc-install:
	@echo "Installing sqlc (v2)..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

sqlc-generate:
	@echo "Generating Go code from SQL (sqlc v2)..."
	@sqlc generate --file db/sqlc.yaml

# ======== MIGRATIONS ========

migrate-install:
	@echo "Installing migrate CLI..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

create-migration:
	@echo "Creating new migration..."
	@migrate create -seq -ext sql -dir ./migrations $(name)
	@echo "Migration '$(name)' created successfully."
