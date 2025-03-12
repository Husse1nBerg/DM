### Generate docs
```bash
swag init -d ./cmd/api/,./internal/server/handlers -o ./docs -g main.go --parseInternal
```

## Running Goose Migrations

### Prerequisites

Before running Goose migrations, ensure you have the following installed and configured:

- **Goose**: Migration tool for PostgreSQL.
  - Install it via:
    ```bash
    go install github.com/pressly/goose/v3@latest
    ```
  - In Mac:
    ```bash
    brew install goose
    ```

- **Docker** (for local setup):
  - Docker should be installed to run the PostgreSQL container.

    ```bash
    docker-compose up -d
    ```

### Apply Migrations

- **Run Goose Migrations**:

    - Ensure .env variables are loaded into the shell:
    ```bash
    source .env
    ```

    - Run goose migrations locally:
    ```bash
    goose -dir db/migrations postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" up
    ```

    - Run goose migrations in production:
    ```bash
    goose -dir db/migrations postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" up
    ```