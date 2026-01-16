package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"hitalent/internal/chat"
	"hitalent/internal/httpapi"
	"hitalent/internal/storage"
)

func main() {
	// Общий логгер
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Контекст для старта
	ctx := context.Background()

	// Подключаемся к Postgres через GORM
	gdb, sqlDB, err := storage.OpenPostgres(ctx)
	if err != nil {
		log.Error("failed to connect postgres", "err", err)
		os.Exit(1)
	}
	defer func() { _ = sqlDB.Close() }()
	log.Info("connected to postgres")

	// Собираем зависимости (repo  service  api  router)
	repo := chat.NewRepo(gdb)
	svc := chat.NewService(repo)
	api := httpapi.NewAPI(svc)

	router := httpapi.NewRouter(api)

	// Middleware
	// RecoverMiddleware ловит панику внутри обработчиков
	//LoggingMiddleware логирует каждый запрос:
	handler := httpapi.RecoverMiddleware(log, router)
	handler = httpapi.LoggingMiddleware(log, handler)

	// HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Info("server started", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error", "err", err)
		os.Exit(1)
	}
}
