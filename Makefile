.PHONY: build test test-v run-docs-gen help

# Default target
build:
	go build -o yoto main.go

# Run all tests
test:
	go test ./...

# Run all tests with verbose output
test-v:
	go test -v ./...

# Generate CLI documentation
docs: build
	mkdir -p docs/commands
	go run cmd/gen-docs/main.go

# Install the binary
install: build
	sudo mv yoto /usr/local/bin/

help:
	@echo "Available targets:"
	@echo "  build     - Build the yoto binary"
	@echo "  test      - Run all tests"
	@echo "  test-v    - Run all tests (verbose)"
	@echo "  docs      - Generate markdown documentation for all commands"
	@echo "  install   - Install the binary to /usr/local/bin"
