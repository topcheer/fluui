# Fluui Makefile — common development tasks
# Usage: make <target>

GOCACHE ?= /tmp/go-cache-fluui
GOFLAGS := -race -short -count=1
TIMEOUT := 300s
BENCHFLAGS := -bench=. -benchmem -benchtime=1s -run=^$

.PHONY: all build vet test test-race bench coverage lint clean help

## all: Build and run all tests (default)
all: build vet test

## build: Compile all packages
build:
	GOCACHE=$(GOCACHE) go build ./...

## vet: Run go vet on all packages
vet:
	go vet ./...

## test: Run short tests without race detector
test:
	GOCACHE=$(GOCACHE) go test -short -count=1 ./... -timeout $(TIMEOUT)

## test-race: Run tests with race detector (recommended)
test-race:
	GOCACHE=$(GOCACHE) go test $(GOFLAGS) ./... -timeout $(TIMEOUT)

## bench: Run all benchmarks
bench:
	GOCACHE=$(GOCACHE) go test $(BENCHFLAGS) ./component/ ./markdown/ ./internal/buffer/ ./render/ ./block/

## coverage: Generate coverage report
coverage:
	GOCACHE=$(GOCACHE) go test -short -count=1 -coverprofile=cov.out ./component/ ./block/ ./app/ ./markdown/ ./render/ ./internal/buffer/ ./internal/term/ ./theme/ ./ai/
	go tool cover -func=cov.out | awk '$$3+0 < 80 && $$3 != "(statements)"'
	@echo "---"
	@go tool cover -func=cov.out | tail -1

## coverage-html: Generate HTML coverage report
coverage-html:
	GOCACHE=$(GOCACHE) go test -short -count=1 -coverprofile=cov.out ./component/ ./block/ ./app/ ./markdown/
	go tool cover -html=cov.out

## lint: Run golangci-lint (if installed)
lint:
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./... ; \
	else \
		echo "golangci-lint not installed, running go vet instead" ; \
		go vet ./... ; \
	fi

## clean: Remove build artifacts and cache
clean:
	rm -f cov.out fluui ai-chat-demo
	rm -rf $(GOCACHE)

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | sed 's/:/\t/'

## install: Install the fluui demo binary
install:
	GOCACHE=$(GOCACHE) go install ./cmd/fluui-demo 2>/dev/null || echo "no demo cmd found"
