.PHONY: all build test fmt vet lint clean

all: fmt vet lint test build

build:
	go build -o bin/habits main.go

test:
	go test -v -cover ./...

fmt:
	go fmt ./...

# Static analysis
vet:
	go vet ./...

# Basic linter using 'go install golang.org/x/lint/golint@latest'
lint:
	golint ./... || true

clean:
	rm -rf bin
