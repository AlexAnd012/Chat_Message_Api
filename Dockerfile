# собираем бинарник
FROM golang:1.24.7-alpine AS builder


WORKDIR /app

# Кеширование зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Устанавливаем goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Копируем весь проект в контейнер
COPY . .
# Собираем бинарник API
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd



# Образ для запуска
FROM alpine:3.20

WORKDIR /app

# Копируем собранный бинарник в runtime образ
COPY --from=builder /out/api /app/api

# Копируем goose в runtime образ
COPY --from=builder /go/bin/goose /app/goose

# Копируем SQL миграции в контейнер
COPY migrations /app/migrations

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/api"]
