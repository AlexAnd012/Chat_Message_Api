package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"hitalent/internal/chat"
)

type API struct {
	svc *chat.Service
}

func NewAPI(svc *chat.Service) *API {
	return &API{svc: svc}
}

// CreateChat POST /chats/
func (a *API) CreateChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	// decodeJSON функция из json.go читает json из r.Body, парсит в req, иначе дает ошибку
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	// вызываем сервис
	c, err := a.svc.CreateChat(r.Context(), req.Title)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

// CreateMessage POST /chats/{id}/messages/
func (a *API) CreateMessage(w http.ResponseWriter, r *http.Request, chatID int64) {
	var req struct {
		Text string `json:"text"`
	}
	// decodeJSON функция из json.go читает json из r.Body, парсит в req, иначе дает ошибку
	if err := decodeJSON(w, r, &req); err != nil {
		return
	}
	// вызываем сервис
	m, err := a.svc.CreateMessage(r.Context(), chatID, req.Text)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, m)
}

// GetChat GET /chats/{id}?limit=N
func (a *API) GetChat(w http.ResponseWriter, r *http.Request, chatID int64) {
	limit := 0
	// Читаем query параметры, если не число, возвращаем ошибку
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = n
	}

	// вызываем сервис, он проверит данные, проверит что чат существует и вернет последние limit сообщения, иначе ошибку
	c, msgs, err := a.svc.GetChatWithMessages(r.Context(), chatID, limit)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	// формируем ответ и отдаем json
	resp := struct {
		Chat     *chat.Chat     `json:"chat"`
		Messages []chat.Message `json:"messages"`
	}{
		Chat:     c,
		Messages: msgs,
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteChat DELETE /chats/{id} возвращает 204
func (a *API) DeleteChat(w http.ResponseWriter, r *http.Request, chatID int64) {
	// вызываем сервис
	if err := a.svc.DeleteChat(r.Context(), chatID); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Вспомогательная функция для перевода доменных ошибок в http статусы
func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, chat.ErrValidation):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, chat.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
