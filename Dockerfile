# syntax=docker/dockerfile:1

## Fetch swagger-ui assets in a node stage, copy into source tree, then build with Go so go:embed includes them
FROM node:18-alpine AS swagger
WORKDIR /tmp
RUN npm init -y
RUN npm install --no-save swagger-ui-dist@latest

FROM golang:1.24 AS builder
WORKDIR /src

# copy go mod first for caching
COPY go.mod go.sum ./
RUN go mod download

# copy the rest of the source
COPY . .

# copy swagger-ui dist from the node stage into the package that uses go:embed
COPY --from=swagger /tmp/node_modules/swagger-ui-dist ./internal/transport/http/swagger-ui

# Ensure project-specific overrides (like swagger-initializer.js) replace the dist's default
# Copy the initializer from the repo into the swagger-ui directory after the dist copy.
COPY internal/transport/http/swagger-ui/swagger-initializer.js ./internal/transport/http/swagger-ui/swagger-initializer.js

# build the binary (go:embed will pick up files under internal/transport/http/swagger-ui)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/api /app/api

EXPOSE 9090

USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
