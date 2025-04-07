#!/bin/sh
set -e

# Check for required environment variables
# if [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_NAME" ]; then
#   echo "Error: Missing required database environment variables"
#   echo "Required: DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME"
#   exit 1
# fi

# # Run migrations
# echo "Running database migrations..."
# cd /app/db/migrations && goose postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE:-disable}" up

# Start the application
echo "Starting application..."
cd /app && exec ./main 