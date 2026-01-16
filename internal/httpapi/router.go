package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

// Handler Интерфейс для удобства тестирования
type Handler interface {
	CreateChat(w http.ResponseWriter, r *http.Request)
	CreateMessage(w http.ResponseWriter, r *http.Request, chatID int64)
	GetChat(w http.ResponseWriter, r *http.Request, chatID int64)
	DeleteChat(w http.ResponseWriter, r *http.Request, chatID int64)
}

// NewRouter используем стандартный роутер из Go и будем матчить пути по префиксу или точному совпадению
func NewRouter(h Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/chats", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chats" {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodPost:
			h.CreateChat(w, r)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})

	// обработчик для /chats/{id}  и  /chats/{id}/messages
	mux.HandleFunc("/chats/", func(w http.ResponseWriter, r *http.Request) {

		path := strings.TrimPrefix(r.URL.Path, "/chats/")
		path = strings.Trim(path, "/")

		// если /chats/ обрабатываем как создание чата
		if path == "" {
			if r.Method == http.MethodPost {
				h.CreateChat(w, r)
				return
			}
			http.NotFound(w, r)
			return
		}

		parts := strings.Split(path, "/")

		if len(parts) == 0 || parts[0] == "" {
			http.NotFound(w, r)
			return
		}

		// Парсим id и проверяем является ли числом
		chatID, ok := parseInt64(parts[0])
		if !ok || chatID <= 0 {
			http.NotFound(w, r)
			return
		}

		// /chats/{id}
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.GetChat(w, r, chatID)
				return
			case http.MethodDelete:
				h.DeleteChat(w, r, chatID)
				return
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		}

		// /chats/{id}/messages
		if len(parts) == 2 && parts[1] == "messages" {
			switch r.Method {
			case http.MethodPost:
				h.CreateMessage(w, r, chatID)
				return
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		}

		http.NotFound(w, r)
	})

	return mux
}

// Функция для обработки ошибок при переводе id
func parseInt64(s string) (int64, bool) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}
