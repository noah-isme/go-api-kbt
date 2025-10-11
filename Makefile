.PHONY: run lint test build migrate docs-install

run:
	go run ./cmd/api

lint:
	golangci-lint run ./...

test:
	go test ./...

build:
	go build ./...

smoke:
	./scripts/smoke.sh

migrate:
	migrate -path migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSLMODE" up

# Download and install Swagger UI static assets into internal/transport/http/swagger-ui
# Requires: node >= 14 and npm
docs-install:
	@echo "Installing swagger-ui-dist and copying assets to internal/transport/http/swagger-ui"
	@mkdir -p internal/transport/http/swagger-ui
	@npm install --no-save swagger-ui-dist@latest
	@cp -r node_modules/swagger-ui-dist/* internal/transport/http/swagger-ui/
	@rm -rf node_modules package-lock.json
	@echo "Swagger UI assets copied. Start the server and visit http://localhost:9090/docs/"
