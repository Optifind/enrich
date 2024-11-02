BIN_DIR := bin
$(BIN_DIR):
	mkdir -p $(BIN_DIR)


# Single run scripts with large database operations
.PHONY: init cluster

HOST=localhost
USERNAME=enrich
DB_NAME=enrich_test
DUMP_FILEPATH=data/db_dumps/latest.dump
HEROKU_APP_NAME=enrich-optifind-v0

# Initialization target
init: $(BIN_DIR)/init_db
	$(BIN_DIR)/init_db
	rm $(DUMP_FILEPATH)
	pg_dump -Fc --no-acl --no-owner -h $(HOST) -U $(USERNAME) -d $(DB_NAME) -f $(DUMP_FILEPATH)

# Build init_db
$(BIN_DIR)/init_db: scripts/init_db/init_db.go $(shell find pkg -type f) | $(BIN_DIR)
	$(GO_BUILD) -o $(BIN_DIR)/init_db scripts/init_db/init_db.go

# Cluster target
cluster: $(BIN_DIR)/cluster
	$(BIN_DIR)/cluster
	rm $(DUMP_FILEPATH)
	pg_dump -Fc --no-acl --no-owner -h $(HOST) -U $(USERNAME) -d $(DB_NAME) -f $(DUMP_FILEPATH)

# Build cluster
$(BIN_DIR)/cluster: scripts/cluster/cluster_products.go $(shell find pkg -type f) | $(BIN_DIR)
	$(GO_BUILD) -o $(BIN_DIR)/cluster scripts/cluster/cluster_products.go


# Runs docker compose that spins up container with "optifind-api" image from Docker Hub
# Image need to be pushed to the repository before running this
compose-up:
	@docker compose -f ./deploy/compose.yaml up -d

compose-down: clean-api


# Product api that runs continuously in a container
.PHONY: run-api stop-api clean-api log-api deploy-api

API_SRC_DIR := cmd/api

# Build api container image
build-api: $(wildcard $(API_SRC_DIR)/*.go) $(wildcard pkg/**/*.go) | $(BIN_DIR)
	@docker build -t lattots/optifind-api -f ./build/api/Dockerfile .

# Start api container with built image
run-api: build-api
	@docker run -d --network="host" --name optifind-api lattots/optifind-api

# Stop running api container
stop-api:
	@docker stop optifind-api

# Remove api container
clean-api: stop-api
	@docker rm optifind-api

log-api:
	@docker logs optifind-api

# Push built api container image to Docker Hub
deploy-api: build-api
	@docker push lattots/optifind-api:latest


.PHONY: test

# Test target
test:
	set -a && . "data/secrets.env" && set +a && \
	$(GO_TEST) -v ./pkg/vector/ && \
	$(GO_TEST) -v ./pkg/langmod/ && \
	$(GO_TEST) -v ./pkg/catalog/

