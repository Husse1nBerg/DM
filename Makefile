# Simple Makefile for a Go project
# ToDo: update the Makefile
ARTIFACT_NAME := dm-web

# Include .env file
include .env
export

# Build the application
all: build

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Create DB container
docker-run:
	@if docker compose up 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

instal_sqlc :
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

install_goose :
	go install github.com/pressly/goose/v3/cmd/goose@latest

swag :
	swag init -g ./cmd/api/main.go -o ./docs

create-migration :
	cd db/migrations && goose create $(name) sql

# Goose migration UP
goose-up:
	cd db/migrations && goose postgres postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME} up

# Goose migration DOWN
goose-down:
	cd db/migrations && goose postgres postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME} down

sqlc:
	sqlc generate
# Test the application
test:
	@echo "Testing..."
	@go test ./tests -v

go-test:
	@go test -v $(shell go list ./... | grep -v /test/)

go-test-with-cover:
	@go test -coverprofile cover.out -v $(shell go list ./... | grep -v /test/)
	@go tool cover -html=cover.out

generate-mocks:
	@mockery --all --with-expecter --keeptree

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
	    air; \
	    echo "Watching...";\
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
	    if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
	        go install github.com/air-verse/air@latest; \
	        air; \
	        echo "Watching...";\
	    else \
	        echo "You chose not to install air. Exiting..."; \
	        exit 1; \
	    fi; \
	fi

.PHONY: all build run test clean
