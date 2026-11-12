# Команды для Go-микросервисов (services/). Фронтенды — через pnpm (см. package.json).
SERVICES := gateway auth users bot
GO_PKGS  := $(foreach m,pkg $(SERVICES),./$(m)/...)

.PHONY: help build test vet fmt tidy lint up down dev-up

help: ## Показать список команд
	@grep -hE '^[a-zA-Z_%-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

dev-%: ## Запустить сервис локально: make dev-gateway | dev-auth | dev-users | dev-bot
	cd services && go run ./$*/cmd/$*

build: ## Собрать все сервисы в services/tmp/
	cd services && for s in $(SERVICES); do go build -o tmp/$$s ./$$s/cmd/$$s || exit 1; done

test: ## Запустить Go-тесты
	cd services && go test $(GO_PKGS)

vet: ## go vet по всем модулям
	cd services && go vet $(GO_PKGS)

fmt: ## gofmt -w
	cd services && gofmt -w .

tidy: ## go mod tidy во всех модулях
	cd services && for m in pkg $(SERVICES); do (cd $$m && go mod tidy) || exit 1; done

lint: vet ## Проверки Go + Biome для фронтендов
	@test -z "$$(cd services && gofmt -l .)" || (echo "gofmt: run 'make fmt'" && cd services && gofmt -l . && exit 1)
	pnpm lint

up: ## Поднять весь стек в Docker (prod-сборка)
	docker compose up -d --build

dev-up: ## Поднять весь стек в Docker с hot-reload
	docker compose -f docker-compose.dev.yml up --build

down: ## Остановить стек
	docker compose down
