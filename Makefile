.PHONY: build test lint clean install run help

# Build variables
BINARY := agentsentinel
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/plexusone/agentsentinel/cmd.Version=$(VERSION)"

## build: Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY) .

## test: Run tests
test:
	go test -v ./...

## test-cover: Run tests with coverage
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	golangci-lint run

## clean: Remove build artifacts
clean:
	rm -f $(BINARY) coverage.out coverage.html

## install: Install to GOPATH/bin
install:
	go install $(LDFLAGS) .

## run: Build and run with watch command
run: build
	./$(BINARY) watch

## run-dry: Build and run in dry-run mode
run-dry: build
	./$(BINARY) watch --dry-run -v

## status: Build and show tmux status
status: build
	./$(BINARY) status

## fmt: Format code
fmt:
	go fmt ./...

## tidy: Tidy go modules
tidy:
	go mod tidy

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'
