.PHONY: all build test fmt vet lint clean

# Run everything
all: fmt vet lint test build

# Build the binary
build:
	go build -o bin/habits main.go

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Static analysis
vet:
	go vet ./...

# Basic linter using 'go install golang.org/x/lint/golint@latest'
lint:
	golint ./... || true

# Clean build artifacts
clean:
	rm -rf bin
