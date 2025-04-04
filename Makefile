SHELL := /bin/bash
include .env
export
export APP_NAME := $(basename $(notdir $(shell pwd)))

.PHONY: help
help: ## display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PNONY: tool
tool: ## install tools
	@go install github.com/xo/xo@latest
	@go install github.com/volatiletech/sqlboiler/v4@latest
	@go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest

.PHONY: up
up: ## docker compose up with air hot reload
	@docker compose --project-name postgres --file ./.docker/compose.yaml up -d

.PHONY: down
down: ## docker compose down
	@docker compose --project-name postgres down --volumes

.PHONY: psql
psql:
	@docker exec -it postgres psql -U postgres

.PHONY: migrate
migrate: ## migrate prisma schema
	@(cd schema && bun run prisma db push)

.PHONY: sync
sync: ## import prisma schema
	@(cd schema && bun run prisma db pull)

.PHONY: schemafmt
schemafmt: 
	@(cd schema && bun run prisma format)

.PHONY: gen
gen: ## generate code
	@go tool github.com/xo/xo schema postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable -o pkg/xogen
	@go tool github.com/stephenafamo/bob/gen/bobgen-psql
	@go tool github.com/volatiletech/sqlboiler/v4 psql
	@go mod tidy

.PHONY: run
run: ## run application
	@go run main.go
