BIN_DIR ?= $(CURDIR)/bin

.PHONY: build
build:
	go build -v -o $(BIN_DIR)/client ./cmd/client

build-tools:
	go build -v -o $(BIN_DIR)/scraper ./cmd/scraper
	go build -v -o $(BIN_DIR)/crypt ./cmd/crypt

.PHONY: test
test:
	go test ./...
