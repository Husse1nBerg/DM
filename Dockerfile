# ---------- BUILD STAGE ----------
FROM golang:1.23-alpine AS build

RUN apk add --no-cache curl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go
# RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# ---------- RUNTIME STAGE ----------
FROM alpine:3.20.1 AS prod

WORKDIR /app

# Copy built binary
COPY --from=build /app/main /app/main

# Optional goose & migration support
# COPY --from=build /go/bin/goose /usr/local/bin/goose
# COPY --from=build /app/db/migrations /app/db/migrations

# Copy entrypoint script
COPY --from=build /app/scripts/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Copy .env file if it exists in context (useful during GitHub Actions build)
COPY .env .env

# Install `bash` and `envsubst` for env file parsing if needed
RUN apk add --no-cache bash

# Export env vars from .env at runtime (sourced inside entrypoint)
# Note: Docker doesn't evaluate ENV from a file directly

# Expose port at runtime
EXPOSE 8080

# Run via entrypoint which will source the .env
ENTRYPOINT ["/app/entrypoint.sh"]
