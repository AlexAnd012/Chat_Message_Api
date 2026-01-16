package chat

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

// CreateChat создаем чат и сохраняем в бд
func (r *Repo) CreateChat(ctx context.Context, title string) (*Chat, error) {
	c := &Chat{Title: title}

	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	return c, nil
}

// GetChatByID возвращаем чат по id или ErrNotFound
func (r *Repo) GetChatByID(ctx context.Context, id int64) (*Chat, error) {
	var c Chat
	err := r.db.WithContext(ctx).
		First(&c, "id = ?", id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get chat by id: %w", err)
	}
	return &c, nil
}

// CreateMessage создаем сообщение в чате
func (r *Repo) CreateMessage(ctx context.Context, chatID int64, text string) (*Message, error) {
	m := &Message{
		ChatID: chatID,
		Text:   text,
	}

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}
	return m, nil
}

// ListLastMessages возвращает последние limit сообщений в чате.
// Используем вспомогательный индекс
func (r *Repo) ListLastMessages(ctx context.Context, chatID int64, limit int) ([]Message, error) {
	var msgs []Message
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("created_at DESC").
		Limit(limit).
		Find(&msgs).
		Error
	if err != nil {
		return nil, fmt.Errorf("list last messages: %w", err)
	}
	return msgs, nil
}

// DeleteChat удаляет чат по id, сообщения удаляются каскадно на уровне БД
func (r *Repo) DeleteChat(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&Chat{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("delete chat: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
