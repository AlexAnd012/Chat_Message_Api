# Chat API (test task)

Простой REST API для чатов и сообщений.

## Функциональность

### Модели
- **Chat**
    - `id: int`
    - `title: string` (1..200, не пустой)
    - `created_at: datetime`
- **Message**
    - `id: int`
    - `chat_id: int` 
    - `text: string` (1..5000, не пустой)
    - `created_at: datetime`

Связь: `Chat 1 — N Message`

### Методы API
- `POST /chats/` — создать чат  
  Body: `{ "title": "..." }`  
  Response: созданный чат

- `POST /chats/{id}/messages/` — отправить сообщение в чат  
  Body: `{ "text": "..." }`  
  Response: созданное сообщение

- `GET /chats/{id}?limit=N` — получить чат и последние N сообщений  
  Query: `limit` (по умолчанию 20, максимум 100)  
  Response: `{ "chat": {...}, "messages": [...] }`  
  `messages` отсортированы по `created_at`

- `DELETE /chats/{id}` — удалить чат и все сообщения  
  Response: `204 No Content`

### Логика и ограничения
- Нельзя отправить сообщение в несуществующий чат `404`.
- Валидация:
    - `title`: trim + длина 1..200
    - `text`: trim + длина 1..5000
- При удалении чата сообщения удаляются каскадно на уровне БД (`ON DELETE CASCADE`).

## Технологии
- Go + `net/http`
- PostgreSQL
- GORM
- Миграции: `goose`
- Docker + docker-compose
- Тесты: `httptest` + `testify`

## Переменные окружения

PORT - порт HTTP сервера
DATABASE_DSN - DSN PostgreSQL

## Структура проекта

hitalent/  
├── cmd/  
│   └── main.go                   # сборка зависимостей, запуск HTTP-сервера  
├── internal/  
│   ├── chat/  
│   │   ├── models.go             # модели Chat/Message + normalize/validate  
│   │   ├── errors.go             # доменные ошибки (ErrValidation, ErrNotFound)  
│   │   ├── repo.go               # репозиторий (GORM), CRUD для чатов/сообщений  
│   │   └── service.go            # бизнес-логика валидация, not found, limit  
│   ├── httpapi/  
│   │   ├── router.go             # роутинг на net/http   
│   │   ├── api.go                # HTTP handlers (CreateChat/CreateMessage/GetChat/DeleteChat)  
│   │   ├── json.go               # decodeJSON/writeJSON/writeError   
│   │   └── middleware.go         # middleware, recover + logging   
│   └── storage/  
│       └── postgres.go           # подключение к PostgreSQL через GORM + настройки пула соединений  
├── migrations/  
│   └── 00001_init.sql            # goose миграция: таблицы chats и messages , каскадное удаление   
├── tests/  
│   └── http_test.go              # тесты API   
├── Dockerfile                       
├── docker-compose.yml            # сервисы db, migrate, api  
├── Makefile                      # команды: up/down/logs/test/migrate-up  
├── go.mod  
├── go.sum  
└── README.md                      
  

## Запуск через Docker Compose  
  
### Требования  
- установлен **Docker**  
- установлен **Make**  

### 1) Поднять сервисы и применить миграции  
Из корня проекта:  

`make up`
`docker compose up -d --build`  
  
### 2) Остановить и удалить контейнеры + volume с БД  

`make down`  
`docker compose down -v`  

### 3) Смотреть логи

`make logs`
`docker compose logs -f --tail=200`

### 4) Запустить тесты ( поднимается httptest сервер и выполняются HTTP запросы )

`make test`
`go test ./... -v`

### 5) Запустить миграции вручную

`make migrate-up`
`docker compose run --rm migrate`
