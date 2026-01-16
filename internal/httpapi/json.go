package httpapi

import (
	"encoding/json"
	"net/http"
)

// Лимит размера тела запроса 1 МБ
const maxBodyBytes = 1 << 20

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	// Создаём JSON декодер и запрещаем неизвестные поля
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// декодируем json в dst
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return err
	}

	return nil
}

// Ставим заголовок ответа и http статус
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Форматируем ошибку
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
