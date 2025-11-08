# Project configuration
PROJECT_NAME := unusedfunc
BINARY_NAME := unusedfunc
PACKAGE := github.com/715d/unusedfunc
CMD_PACKAGE := $(PACKAGE)/cmd/unusedfunc

# Build configuration
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT) -w -s"

# Directories
BUILD_DIR := build
TESTDATA_DIR := testdata

# Default target
.DEFAULT_GOAL := help

##@ General

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <target>\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 } /^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Download and verify dependencies
	@go mod download
	@go mod verify
	@go mod tidy

.PHONY: tools
tools: ## Install development tools
	@go install github.com/goreleaser/goreleaser/v2@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest

.PHONY: fmt
fmt: ## Format code with gofmt and goimports
	@goimports -w -local $(PACKAGE) .

.PHONY: lint
lint: ## Run linters
	@golangci-lint run ./...

.PHONY: check
check: format lint ## Run all code quality checks

.PHONY: build
build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PACKAGE)

.PHONY: test
test: ## Run tests with coverage
	@mkdir -p $(BUILD_DIR)
	@go test -race -coverprofile=$(BUILD_DIR)/coverage.out ./...
	@go tool cover -func=$(BUILD_DIR)/coverage.out | tail -1

.PHONY: clean
clean: ## Clean build artifacts
	@rm -rf $(BUILD_DIR)

