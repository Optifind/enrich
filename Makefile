.PHONY: init api test build-init build-api all clean

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
	rm data/db_dumps/latest.dump
	pg_dump -Fc --no-acl --no-owner -h localhost -U enrich -d enrich_test -f data/db_dumps/latest.dump

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
$(BINARY_DIR)/init_db: scripts/init_db/init_db.go $(shell find pkg -type f) | $(BINARY_DIR)
	$(GO_BUILD) -o $(BINARY_DIR)/init_db scripts/init_db/init_db.go

# Build api
$(BINARY_DIR)/api: cmd/api/api.go $(shell find pkg -type f) | $(BINARY_DIR)
	$(GO_BUILD) -o $(BINARY_DIR)/api cmd/api/api.go

# Clean build artifacts
clean:
	rm -rf $(BINARY_DIR)
