# Makefile for Minishell Tester

.PHONY: all build clean test test-quiet list create-tests

# Go compiler
GO := go

# Binary name
BIN := maybe

# Build flags
BUILD_FLAGS := -ldflags="-s -w"

# Source files
SRC := main.go test-loader.go go-minishell-tester-core.go

all: build

build:
	@echo "Building $(BIN)..."
	@$(GO) build $(BUILD_FLAGS) -o $(BIN) $(SRC)
	@echo "Build complete. Run './$(BIN) --help' for usage."

clean:
	@echo "Cleaning up..."
	@rm -f $(BIN)
	@rm -rf outfiles mini_outfiles bash_outfiles
	@echo "Clean complete."

test: build
	@echo "Running all tests..."
	@./$(BIN)

test-quiet: build
	@echo "Running all tests (quiet mode)..."
	@./$(BIN) --no-details

list: build
	@./$(BIN) --list

create-tests: build
	@echo "Creating default test files..."
	@./$(BIN) --create-tests