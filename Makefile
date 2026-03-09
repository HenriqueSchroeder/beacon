.PHONY: build test coverage lint clean run help

BIN_NAME=beacon
BIN_DIR=bin
CMD_DIR=cmd/$(BIN_NAME)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_FLAGS=-ldflags "-X main.version=$(VERSION)"

help:
	@echo "Beacon Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build      - Build the binary"
	@echo "  make run        - Run the CLI"
	@echo "  make test       - Run tests"
	@echo "  make lint       - Run Go linter"
	@echo "  make clean      - Remove build artifacts"

build:
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BIN_NAME) ./$(CMD_DIR)
	@echo "✓ Built: $(BIN_DIR)/$(BIN_NAME)"

run: build
	./$(BIN_DIR)/$(BIN_NAME)

test:
	go test -v -cover ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm -f coverage.out

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BIN_DIR)
	go clean

install: build
	go install $(BUILD_FLAGS) ./$(CMD_DIR)
	@echo "✓ Installed: $(BIN_NAME)"
