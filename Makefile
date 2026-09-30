.PHONY: build test lint install clean build-mcp install-mcp build-all

BIN_EXT := $(if $(filter windows,$(shell go env GOOS)),.exe,)

build:
	go build -trimpath -o bin/ankiweb$(BIN_EXT) ./cmd/ankiweb

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/ankiweb

clean:
	rm -rf bin/

build-mcp:
	go build -trimpath -o bin/ankiweb-mcp$(BIN_EXT) ./cmd/ankiweb-mcp

install-mcp:
	go install ./cmd/ankiweb-mcp

build-all: build build-mcp
