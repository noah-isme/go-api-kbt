.PHONY: run lint test build migrate

run:
go run ./cmd/api

lint:
golangci-lint run ./...

test:
go test ./...

build:
go build ./...

migrate:
migrate -path migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSLMODE" up
