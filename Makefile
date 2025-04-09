# DM-WEB Backend Makefile
ARTIFACT_NAME := dm-web

# Include .env file
include .env
export

####################
# APPLICATION COMMANDS
####################

# Build targets
.PHONY: all build clean

all: build

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

clean:
	@echo "Cleaning..."
	@rm -f main

# Run targets
.PHONY: run watch

run:
	@go run cmd/api/main.go

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

####################
# TOOLS INSTALLATION
####################
.PHONY: install-sqlc install-goose install-tools

install-sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@latest

install-mockery:
	go install github.com/vektra/mockery/v2@latest

install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

install-tools: install-sqlc install-goose install-mockery install-swag

####################
# CODE GENERATION
####################
.PHONY: swag sqlc generate-mocks

swag:
	swag init -g ./cmd/api/main.go -o ./docs

sqlc:
	sqlc generate

generate-mocks:
	@mockery --all --with-expecter --keeptree

####################
# DOCKER COMMANDS
####################
.PHONY: dev-up dev-down local-up local-down

# Development environment (DB only)
dev-up:
	docker compose -f docker-compose.dev.yml up -d

dev-down:
	docker compose -f docker-compose.dev.yml down

# Local environment (app + DB)
local-up:
	docker compose -f docker-compose.local.yml up -d

local-down:
	docker compose -f docker-compose.local.yml down

####################
# DATABASE COMMANDS
####################
.PHONY: create-migration goose-up goose-down db-seed

create-migration:
	cd db/migrations && goose create $(name) sql

goose-up:
	@echo "Upgrading production/development database..."
	cd db/migrations && goose postgres postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable up

goose-down:
	@echo "Downgrading production/development database..."
	cd db/migrations && goose postgres postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable down

db-seed:
	@echo "Seeding production/development database with initial data..."
	@go run scripts/prod_seed/main.go

####################
# TEST DATABASE COMMANDS
####################
.PHONY: test-db-up test-db-down test-goose-up test-goose-down test-db-seed test-db-reset 

test-db-up:
	@echo "Starting test database..."
	docker compose -f docker-compose.dev.yml up -d test_db

test-db-down:
	@echo "Stopping test database..."
	docker compose -f docker-compose.dev.yml stop test_db
	@echo "Removing test database..."
	docker compose -f docker-compose.dev.yml rm -f test_db

test-goose-up:
	@echo "Upgrading test database..."
	cd db/migrations && goose postgres postgres://${TEST_DB_USER}:${TEST_DB_PASSWORD}@${TEST_DB_HOST}:${TEST_DB_PORT}/${TEST_DB_NAME}?sslmode=disable up

test-goose-down:
	@echo "Downgrading test database..."
	cd db/migrations && goose postgres postgres://${TEST_DB_USER}:${TEST_DB_PASSWORD}@${TEST_DB_HOST}:${TEST_DB_PORT}/${TEST_DB_NAME}?sslmode=disable down

test-db-seed:
	@echo "Seeding test database with test data..."
	@go run scripts/test_seed/main.go

test-db-run: test-db-up test-goose-up

test-db-reset: test-goose-down test-goose-up 

####################
# TESTING COMMANDS
####################
.PHONY: test go-test go-test-with-cover coverage coverage-handlers

test:
	@echo "Testing..."
	@go test ./tests -v

go-test:
	@go test -v $(shell go list ./... | grep -v /tests/)

go-test-with-cover:
	@go test -coverprofile cover.out -v $(shell go list ./... | grep -v /tests/)
	@go tool cover -html=cover.out

coverage:
	@echo "Running comprehensive test coverage..."
	@go test -coverprofile=coverage.out -covermode=atomic ./... 
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-handlers:
	@echo "Running coverage for handlers..."
	@go test -coverprofile=handlers-coverage.out -covermode=atomic ./internal/server/...
	@go tool cover -func=handlers-coverage.out
	@go tool cover -html=handlers-coverage.out -o handlers-coverage.html
	@echo "Handlers coverage report generated: handlers-coverage.html"
