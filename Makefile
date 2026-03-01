.PHONY: build test lint clean run

BINARY = snap

# Build the binary
build:
	go build -o $(BINARY).exe .

# Run all tests
test:
	go test ./... -v

# Run the linter
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -f $(BINARY).exe

# Run the binary (for quick dev testing)
run:
	go run . $(ARGS)
