package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"hitalent/internal/chat"
	"hitalent/internal/httpapi"
	"hitalent/internal/storage"
)

func TestChatAPI_Path_And_CascadeDelete(t *testing.T) {

	// Запускаем тестовый сервер
	srv, db := startTestServer(t)
	defer srv.Close()

	//  Create chat
	chatResp := struct {
		ID        int64     `json:"id"`
		Title     string    `json:"title"`
		CreatedAt time.Time `json:"created_at"`
	}{}

	// делаем Post и проверяем
	status, body := doJSON(t, http.MethodPost, srv.URL+"/chats/", map[string]any{
		"title": "  Test chat  ",
	})

	require.Equal(t, http.StatusCreated, status)
	require.NoError(t, json.Unmarshal(body, &chatResp))
	require.Greater(t, chatResp.ID, int64(0))
	require.Equal(t, "Test chat", chatResp.Title)
	require.False(t, chatResp.CreatedAt.IsZero())

	// Create message
	msgResp := struct {
		ID        int64     `json:"id"`
		ChatID    int64     `json:"chat_id"`
		Text      string    `json:"text"`
		CreatedAt time.Time `json:"created_at"`
	}{}

	// делаем Post и проверяем
	status, body = doJSON(t, http.MethodPost, fmt.Sprintf("%s/chats/%d/messages/", srv.URL, chatResp.ID), map[string]any{
		"text": "  Hello  ",
	})
	require.Equal(t, http.StatusCreated, status)
	require.NoError(t, json.Unmarshal(body, &msgResp))
	require.Greater(t, msgResp.ID, int64(0))
	require.Equal(t, chatResp.ID, msgResp.ChatID)
	require.Equal(t, "Hello", msgResp.Text)
	require.False(t, msgResp.CreatedAt.IsZero())

	//  Get chat with messages
	getResp := struct {
		Chat     map[string]any   `json:"chat"`
		Messages []map[string]any `json:"messages"`
	}{}

	// делаем Get и проверяем
	status, body = doRaw(t, http.MethodGet, fmt.Sprintf("%s/chats/%d?limit=20", srv.URL, chatResp.ID), nil)
	require.Equal(t, http.StatusOK, status)
	require.NoError(t, json.Unmarshal(body, &getResp))
	require.Equal(t, float64(chatResp.ID), getResp.Chat["id"])
	require.Len(t, getResp.Messages, 1)

	// Удаляем чат, ожидаем 204
	status, _ = doRaw(t, http.MethodDelete, fmt.Sprintf("%s/chats/%d", srv.URL, chatResp.ID), nil)
	require.Equal(t, http.StatusNoContent, status)

	// Get удаленного чата, ожидаем 404
	status, _ = doRaw(t, http.MethodGet, fmt.Sprintf("%s/chats/%d", srv.URL, chatResp.ID), nil)
	require.Equal(t, http.StatusNotFound, status)

	// Проверка каскадного удаления
	var cnt int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM messages WHERE chat_id=$1", chatResp.ID).Scan(&cnt)
	require.NoError(t, err)
	require.Equal(t, 0, cnt)
}

// Проверка отправки сообщения в несуществующий чат, ожидаем 404
func TestChatAPI_CreateMessage_NotFound(t *testing.T) {

	srv, _ := startTestServer(t)
	defer srv.Close()

	status, _ := doJSON(t, http.MethodPost, srv.URL+"/chats/999999/messages/", map[string]any{
		"text": "hi",
	})
	require.Equal(t, http.StatusNotFound, status)
}

// Вспомогательные функции для тестов
// поднимаем HTTP-сервер для тестов
func startTestServer(t *testing.T) (*httptest.Server, *sql.DB) {
	t.Helper()

	// Логгер
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Подключаемся к Postgres
	ctx := context.Background()
	gdb, sqlDB, err := storage.OpenPostgres(ctx)
	require.NoError(t, err)

	// Чистим БД перед каждым тестом, чтобы тесты были независимыми
	cleanDB(t, sqlDB)

	// Собираем приложение
	repo := chat.NewRepo(gdb)
	svc := chat.NewService(repo)
	api := httpapi.NewAPI(svc)
	router := httpapi.NewRouter(api)

	// Middleware
	handler := httpapi.RecoverMiddleware(log, router)

	// закрываем sqlDB после завершения теста
	t.Cleanup(func() { _ = sqlDB.Close() })

	return httptest.NewServer(handler), sqlDB
}

// Очищаем таблицы перед тестом
func cleanDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE TABLE messages, chats RESTART IDENTITY CASCADE;`)
	require.NoError(t, err)
}

// Хелпер для JSON запросов
func doJSON(t *testing.T, method, url string, payload any) (int, []byte) {
	t.Helper()

	b, err := json.Marshal(payload)
	require.NoError(t, err)

	// вызываем doRaw и возвращаем HTTP status code и тело ответа в массиве байт
	return doRaw(t, method, url, b)
}

// Отправляем HTTP-запрос на httptest.Server
func doRaw(t *testing.T, method, url string, body []byte) (int, []byte) {
	t.Helper()

	// превращаем []byte в reader
	var rbody *bytes.Reader
	if body == nil {
		rbody = bytes.NewReader(nil)
	} else {
		rbody = bytes.NewReader(body)
	}

	// Создаем Http запрос
	req, err := http.NewRequest(method, url, rbody)
	// остановим тест, если не создастся запрос
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Выполняем запрос
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Читаем тело ответа
	respBytes, err := readAll(resp)
	require.NoError(t, err)
	return resp.StatusCode, respBytes
}

// Функция, которая читает весь resp.Body и возвращает байты
func readAll(resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}
