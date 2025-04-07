FROM golang:1.23-alpine AS build

RUN apk add --no-cache curl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.20.1 AS prod

WORKDIR /app

COPY --from=build /app/main /app/main
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY --from=build /app/db/migrations /app/db/migrations
COPY --from=build /app/scripts/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE ${PORT}
CMD ["/app/entrypoint.sh"]
