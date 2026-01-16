COMPOSE=docker compose

.PHONY: up down logs test migrate-up

# Запускает все сервисы из docker-compose.yml
up:
	$(COMPOSE) up -d --build

# Останавливает и удаляет контейнеры
down:
	$(COMPOSE) down -v

# Показывает логи всех сервисов
logs:
	$(COMPOSE) logs -f --tail=200

# Запускает тесты по всему проекту
test:
	go test ./... -v

# Запускает migrate один раз
migrate-up:
	$(COMPOSE) run --rm migrate

