all: build test

build:
	@echo "Building..."
	
	@go build -o main cmd/api/main.go

migrate:
	@echo "====> Adding a new migration"
	@if [ -z "$(name)" ]; then echo "Migration name is required"; exit 1; fi
	@migrate create -ext sql -dir internal/infra/database/migrate/migrations $(name)

migrate-up: 
	@echo "====> Applying all pending migrations"
	@go run internal/infra/database/migrate/migrate.go up

migrate-down: 
	@echo "====> Reverting all migrations"
	@go run internal/infra/database/migrate/migrate.go down

.PHONY: all build run test clean watch docker-run docker-down itest migrate-up migrate-down

