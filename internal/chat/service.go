package chat

import (
	"context"
	"fmt"
)

// Константы лимитов
const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// CreateChat создаем chat, используем функции для валидации из model.go и вызываем репозиторий
// NormalizeTitle убираем пробелы и переводы строк в заголовке
// ValidateTitle после того как убрали пробелы, проверяем длину заголовка
func (s *Service) CreateChat(ctx context.Context, title string) (*Chat, error) {
	title = NormalizeTitle(title)
	if err := ValidateTitle(title); err != nil {
		return nil, err
	}
	return s.repo.CreateChat(ctx, title)
}

// CreateMessage Создаем message, используем функции для валидации из model.go и вызываем репозиторий
// NormalizeText убираем пробелы и переводы строк в поле текст
// ValidateText после того как убрали пробелы, проверяем длину поля текст
func (s *Service) CreateMessage(ctx context.Context, chatID int64, text string) (*Message, error) {
	text = NormalizeText(text)
	if err := ValidateText(text); err != nil {
		return nil, err
	}
	// Условие по которому нельзя отправить сообщение в несуществующий чат
	if _, err := s.repo.GetChatByID(ctx, chatID); err != nil {
		return nil, err // ErrNotFound уйдёт наверх и превратится в 404 в HTTP
	}
	return s.repo.CreateMessage(ctx, chatID, text)
}

// GetChatWithMessages возвращаем чат и последние limit сообщений, отсортированные по created_at (ASC) и вызываем репозиторий
func (s *Service) GetChatWithMessages(ctx context.Context, chatID int64, limit int) (*Chat, []Message, error) {
	limit, err := normalizeLimit(limit)
	if err != nil {
		return nil, nil, err
	}
	c, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, nil, err
	}
	msgs, err := s.repo.ListLastMessages(ctx, chatID, limit) // приходит DESC
	if err != nil {
		return nil, nil, err
	}
	// по условию сообщения отсортированы по created_at
	// мы берем последние N по DESC и разворачиваем в ASC (быстрее и индексы уже в нужном порядке)
	reverseMessages(msgs)
	return c, msgs, nil
}

// DeleteChat удаляем каскадно
// repo.DeleteChat уже возвращает ErrNotFound если RowsAffected == 0
func (s *Service) DeleteChat(ctx context.Context, chatID int64) error {
	return s.repo.DeleteChat(ctx, chatID)
}

// Валидация лимита
func normalizeLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultLimit, nil
	}
	if limit < 0 {
		return 0, fmt.Errorf("%w: limit must be positive", ErrValidation)
	}
	if limit > maxLimit {
		return maxLimit, nil
	}
	return limit, nil
}

// Вспомогательная функция для разворота из DESC в ASC (разворачиваем слайс)
func reverseMessages(msgs []Message) {
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
}
