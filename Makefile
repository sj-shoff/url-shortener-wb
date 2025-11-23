.PHONY: run build migrate-up migrate-down docker-up docker-down
include .env
export

run:
	go run cmd/url-shortener/main.go

build:
	go build -o bin/service-courier cmd/url-shortener/main.go

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" down