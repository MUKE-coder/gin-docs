.PHONY: build test lint run clean

# Build the demo application
build:
	go build -o bin/gin-docs-demo .

# Run all tests
test:
	go test ./... -v -count=1

# Run tests with coverage
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run go vet
vet:
	go vet ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

# Run the demo application
run:
	go run main.go

# Run the basic example
run-basic:
	go run examples/basic/main.go

# Run the full example
run-full:
	go run examples/full/main.go

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Tidy dependencies
tidy:
	go mod tidy

# Check everything
check: vet test
	@echo "All checks passed!"
