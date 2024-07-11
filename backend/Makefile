.PHONY: init api test build-init build-api all

# Variables
GO_RUN=go run
GO_BUILD=go build
GO_TEST=go test
BINARY_DIR=bin

# Default target
all: init api test

# Initialization target
init: $(BINARY_DIR)/init_db
	$(BINARY_DIR)/init_db

# API target
api: $(BINARY_DIR)/api
	$(BINARY_DIR)/api

# Test target
test:
	$(GO_TEST) ./pkg/vector/

# Build init_db binary
build-init: $(BINARY_DIR)/init_db

# Build api binary
build-api: $(BINARY_DIR)/api

# Ensure binaries directory exists
$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

# Build init_db
$(BINARY_DIR)/init_db: scripts/init_db/init_db.go | $(BINARY_DIR)
	$(GO_BUILD) -o $(BINARY_DIR)/init_db scripts/init_db/init_db.go

# Build api
$(BINARY_DIR)/api: cmd/api/api.go | $(BINARY_DIR)
	$(GO_BUILD) -o $(BINARY_DIR)/api cmd/api/api.go

# Dependencies
scripts/init_db/init_db.go: pkg/catalog/catalog.go pkg/config/config.go pkg/prompts/prompts.go
pkg/catalog/catalog.go: pkg/config/config.go pkg/product/product.go pkg/prompts/prompts.go pkg/vector/vector.go
pkg/product/product.go: pkg/product/util.go pkg/config/config.go pkg/prompts/prompts.go
pkg/handler/handler.go: pkg/config/config.go pkg/catalog/catalog.go pkg/product/product.go

# Ensure dependencies are up-to-date
pkg/handler/handler.go: api
