.PHONY: build test lint install clean

BIN_EXT := $(if $(filter windows,$(shell go env GOOS)),.exe,)

build:
	go build -o bin/pubmed-pp-cli$(BIN_EXT) ./cmd/pubmed-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/pubmed-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/pubmed-pp-mcp$(BIN_EXT) ./cmd/pubmed-pp-mcp

install-mcp:
	go install ./cmd/pubmed-pp-mcp

build-all: build build-mcp
