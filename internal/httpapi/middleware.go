package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

// Структура для отслеживания статус-кода
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware Создаём statusWriter, запоминаем время старта и передаем управление роутеру/хендлерам
// Логируем метод, путь, статус, время обработки
func LoggingMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: 200}
		start := time.Now()

		next.ServeHTTP(sw, r)

		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// RecoverMiddleware Оборачиваем выполнение хендлера в defer и отлавливаем панику
func RecoverMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic recovered", "panic", rec)
				writeError(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
