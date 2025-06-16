# ---------- BUILD STAGE ----------
FROM golang:1.23-alpine AS build

RUN apk add --no-cache curl

WORKDIR /app

# # Install Goose
# RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go

# ---------- RUNTIME STAGE ----------
FROM alpine:3.20.1 AS prod

WORKDIR /app

# Copy built binary
COPY --from=build /app/main /app/main

# # Copy Goose binary
# COPY --from=build /go/bin/goose /usr/local/bin/goose
# # Copy migration files
# COPY --from=build /app/db/migrations /app/db/migrations

# Copy entrypoint script
COPY --from=build /app/scripts/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
