BINARY_NAME=datadog2dynatrace
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/datadog2dynatrace/datadog2dynatrace/internal/config.Version=$(VERSION)"

.PHONY: build clean test lint install

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/datadog2dynatrace

install:
	go install $(LDFLAGS) ./cmd/datadog2dynatrace

test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

all: clean build test
