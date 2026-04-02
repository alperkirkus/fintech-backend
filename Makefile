.PHONY: run build tidy lint db-up db-down migrate-up migrate-down migrate-create

# ── Local dev ──────────────────────────────────────────────────────────────────

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

# ── Database (Docker Compose) ──────────────────────────────────────────────────

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

# ── Migrations (requires golang-migrate CLI) ───────────────────────────────────

migrate-up:
	migrate -path internal/database/migrations \
	        -database "$(shell grep DB_ .env | sed 's/DB_HOST=/host=/;s/DB_PORT=/ port=/;s/DB_NAME=/ dbname=/;s/DB_USER=/ user=/;s/DB_PASSWORD=/ password=/;s/DB_SSL_MODE=/ sslmode=/' | tr '\n' ' ')" \
	        up

migrate-down:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate \
		-path internal/database/migrations \
		-database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" \
		down 1

# Usage: make migrate-create NAME=add_something
migrate-create:
	migrate create -ext sql -dir internal/database/migrations -seq $(NAME)
