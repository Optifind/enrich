.PHONY: init api test

init:
	go run ./scripts/init_db/init_db.go

api:
	go run ./cmd/api/api.go

test:
	go test ./pkg/vector/
