# Welfare Settlement Exception Resolution Platform — Makefile.

APP_MODULE := github.com/welfare/settlement-resolver
APP_BINARY := bin/server
GOFLAGS := -trimpath
WEB_DIR := web
DB_DSN ?= postgres://resolver:resolver@127.0.0.1:5432/welfare_resolver?sslmode=disable

.PHONY: help
help:
	@echo "Welfare Settlement Exception Resolution Platform"
	@echo "Targets:"
	@echo "  build         - Build server binary"
	@echo "  run           - Build and run server"
	@echo "  fmt           - Run gofmt -s on all Go files"
	@echo "  vet           - Run go vet ./..."
	@echo "  test          - Run go test ./..."
	@echo "  test-race     - Run go test -race ./..."
	@echo "  test-integration - Run integration tests (requires live PostgreSQL)"
	@echo "  frontend-install - Install web frontend dependencies"
	@echo "  frontend-test - Run frontend tests"
	@echo "  frontend-build - Build frontend for production"
	@echo "  migrate-up    - Apply migrations"
	@echo "  seed          - Insert demo seed data"
	@echo "  openapi-validate - Validate OpenAPI spec"
	@echo "  clean         - Remove build outputs"

.PHONY: build
build:
	go build $(GOFLAGS) -o $(APP_BINARY) ./cmd/server

.PHONY: run
run: build
	./$(APP_BINARY)

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: test-race
test-race:
	go test -race ./...

.PHONY: test-integration
test-integration:
	go test -tags=integration ./...

.PHONY: frontend-install
frontend-install:
	cd $(WEB_DIR) && npm install

.PHONY: frontend-test
frontend-test:
	cd $(WEB_DIR) && npm run test -- --run

.PHONY: frontend-build
frontend-build:
	cd $(WEB_DIR) && npm run build

.PHONY: migrate-up
migrate-up:
	go run ./cmd/server --migrate-up

.PHONY: seed
seed:
	go run ./cmd/server --seed

.PHONY: clean
clean:
	rm -rf bin $(WEB_DIR)/dist $(WEB_DIR)/node_modules storage
