# Makefile for kube

# Variables
BINARY_NAME=kube
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date +%Y-%m-%d_%T)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# List of all kube-* binaries
KUBE_BINARIES=kube-pods kube-services kube-switch-context kube-switch-namespace kube-logs kube-port-forward kube-exec kube-deploy kube-rollout

# Default target
.PHONY: all
all: build

# Build main binary (from cmd/kube)
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o ${BINARY_NAME} ./cmd/kube

# Build all kube-* binaries from cmd/*
.PHONY: build-all-tools
build-all-tools:
	@echo "Building all kube tools..."
	@for binary in ${KUBE_BINARIES}; do \
		echo "Building $$binary..."; \
		go build ${LDFLAGS} -o $$binary ./cmd/$$binary; \
	done

# Build everything
.PHONY: build-all
build-all: build build-all-tools

# Build for multiple platforms
.PHONY: build-cross-platform
build-cross-platform:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-amd64 ./cmd/kube
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-amd64 ./cmd/kube
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-arm64 ./cmd/kube
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-windows-amd64.exe ./cmd/kube

# Install binary to /usr/local/bin
.PHONY: install
install: build
	@echo "Installing ${BINARY_NAME} to /usr/local/bin..."
	sudo mv ${BINARY_NAME} /usr/local/bin/

# Install all tools to /usr/local/bin
.PHONY: install-all
install-all: build-all
	@echo "Installing all tools to /usr/local/bin..."
	sudo mv ${BINARY_NAME} /usr/local/bin/
	@for binary in ${KUBE_BINARIES}; do \
		echo "Installing $$binary..."; \
		sudo mv $$binary /usr/local/bin/; \
	done

# Uninstall main binary from /usr/local/bin
.PHONY: uninstall
uninstall:
	@echo "Uninstalling ${BINARY_NAME} from /usr/local/bin..."
	sudo rm -f /usr/local/bin/${BINARY_NAME}

# Uninstall all tools from /usr/local/bin
.PHONY: uninstall-all
uninstall-all:
	@echo "Uninstalling all tools from /usr/local/bin..."
	sudo rm -f /usr/local/bin/${BINARY_NAME}
	@for binary in ${KUBE_BINARIES}; do \
		echo "Uninstalling $$binary..."; \
		sudo rm -f /usr/local/bin/$$binary; \
	done

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}-*
	rm -f ${KUBE_BINARIES}
	rm -f coverage.out coverage.html

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download

# Run the application
.PHONY: run
run:
	go run ./cmd/kube $(ARGS)

# Development helpers
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	go mod download

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build               - Build the main binary"
	@echo "  build-all-tools     - Build all kube-* tools"
	@echo "  build-all           - Build main binary and all tools"
	@echo "  build-cross-platform - Build for multiple platforms"
	@echo "  install             - Install main binary to /usr/local/bin"
	@echo "  install-all         - Install all binaries to /usr/local/bin"
	@echo "  uninstall           - Uninstall main binary from /usr/local/bin"
	@echo "  uninstall-all       - Uninstall all binaries from /usr/local/bin"
	@echo "  test                - Run tests"
	@echo "  test-coverage       - Run tests with coverage"
	@echo "  fmt                 - Format code"
	@echo "  lint                - Lint code"
	@echo "  tidy                - Tidy dependencies"
	@echo "  clean               - Clean build artifacts"
	@echo "  deps                - Download dependencies"
	@echo "  run                 - Run the application (use ARGS=... for arguments)"
	@echo "  dev-setup           - Setup development environment"
	@echo "  help                - Show this help"
