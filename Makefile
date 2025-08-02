APP_NAME := habits
VERSION  ?= $(shell git describe --tags --dirty --always)
BUILD    ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS  := -X 'brk3.github.io/habits/cmd.Version=$(VERSION)' -X 'brk3.github.io/habits/cmd.BuildDate=$(BUILD)'

BIN_DIR  := bin

.PHONY: all build test fmt vet lint clean

all: fmt vet lint test build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-linux main.go

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-linux-arm64 main.go

build-macos:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME)-macos main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(APP_NAME).exe main.go

build-all: build-linux build-linux-arm64 build-macos build-windows

clean:
	rm -rf bin dist

test:
	go test -v -cover ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	staticcheck ./...
