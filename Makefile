# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company
# SPDX-License-Identifier: Apache-2.0

# Detect OS and architecture
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Build variables
GO_BUILDFLAGS ?=
GO_LDFLAGS ?=
GO_TESTFLAGS ?=

# Version information
BININFO_VERSION     ?= $(shell git describe --tags --always --abbrev=7 2>/dev/null || echo "dev")
BININFO_COMMIT_HASH ?= $(shell git rev-parse --verify HEAD 2>/dev/null || echo "unknown")
BININFO_BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Binary name and paths
BINARY_NAME = oomkill-exporter
CMD_DIR = cmd/oomkill-exporter
BUILD_DIR = build
DIST_DIR = dist

# Install paths
DESTDIR =
ifeq ($(UNAME_S),Darwin)
	PREFIX = /usr/local
else
	PREFIX = /usr
endif

# Default target
.DEFAULT_GOAL := build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(GO_BUILDFLAGS) \
		-ldflags "-s -w \
			-X github.com/sapcc/go-api-declarations/bininfo.binName=oomkill-exporter \
			-X github.com/sapcc/go-api-declarations/bininfo.version=$(BININFO_VERSION) \
			-X github.com/sapcc/go-api-declarations/bininfo.commit=$(BININFO_COMMIT_HASH) \
			-X github.com/sapcc/go-api-declarations/bininfo.buildDate=$(BININFO_BUILD_DATE) \
			$(GO_LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Install binary
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(DESTDIR)$(PREFIX)/bin..."
	install -d -m 0755 "$(DESTDIR)$(PREFIX)/bin"
	install -m 0755 $(BUILD_DIR)/$(BINARY_NAME) "$(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test $(GO_TESTFLAGS) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	go test -v -coverprofile=$(BUILD_DIR)/coverage.out -covermode=atomic ./...
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

# Run linter
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	fi
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w -local github.com/sapcc/kubernetes-oomkill-exporter .; \
	fi

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build \
		--build-arg BININFO_VERSION=$(BININFO_VERSION) \
		--build-arg BININFO_COMMIT_HASH=$(BININFO_COMMIT_HASH) \
		--build-arg BININFO_BUILD_DATE=$(BININFO_BUILD_DATE) \
		-t $(BINARY_NAME):$(BININFO_VERSION) \
		-t $(BINARY_NAME):latest \
		.

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	go clean

# Run all checks
.PHONY: check
check: lint test
	@echo "All checks passed!"

# Display help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build             Build the binary (default)"
	@echo "  install           Install the binary to $(PREFIX)/bin"
	@echo "  test              Run tests"
	@echo "  test-coverage     Run tests with coverage report"
	@echo "  lint              Run golangci-lint"
	@echo "  fmt               Format code with go fmt and goimports"
	@echo "  tidy              Tidy go.mod dependencies"
	@echo "  docker-build      Build Docker image"
	@echo "  clean             Clean build artifacts"
	@echo "  check             Run lint and tests"
	@echo "  help              Display this help message"
	@echo ""
	@echo "Variables:"
	@echo "  PREFIX=$(PREFIX)"
	@echo "  BININFO_VERSION=$(BININFO_VERSION)"
	@echo "  BININFO_COMMIT_HASH=$(BININFO_COMMIT_HASH)"
	@echo "  BININFO_BUILD_DATE=$(BININFO_BUILD_DATE)"
