#!/bin/sh
set -e

# Run migrations
echo "Running database migrations..."
cd /app/db/migrations && goose postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE:-disable}" up

# Start the application
echo "Starting application..."
cd /app && exec ./main 