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

clean:
	rm -rf bin
