package chat

import (
	"fmt"
	"strings"
	"time"
)

// Chat модель
type Chat struct {
	ID        int64     `gorm:"primaryKey;column:id" json:"id"`
	Title     string    `gorm:"column:title;type:varchar(200);not null" json:"title"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
}

// Message модель
type Message struct {
	ID        int64     `gorm:"primaryKey;column:id" json:"id"`
	ChatID    int64     `gorm:"column:chat_id;not null" json:"chat_id"`
	Text      string    `gorm:"column:text;type:varchar(5000);not null" json:"text"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
}

// NormalizeTitle убираем пробелы и переводы строк в заголовке
func NormalizeTitle(title string) string {
	return strings.TrimSpace(title)
}

// ValidateTitle после того как убрали пробелы, проверяем длину заголовка
func ValidateTitle(title string) error {
	title = NormalizeTitle(title)
	n := len([]rune(title))
	// проверяем длину по рунам если не ASCII символы
	if n < 1 || n > 200 {
		return fmt.Errorf("%w: title length must be 1..200", ErrValidation)
	}
	return nil
}

// NormalizeText убираем пробелы и переводы строк в поле текст
func NormalizeText(text string) string {
	return strings.TrimSpace(text)
}

// ValidateText после того как убрали пробелы, проверяем длину поля текст
func ValidateText(text string) error {
	text = NormalizeText(text)
	// проверяем длину по рунам если не ASCII символы
	n := len([]rune(text))
	if n < 1 || n > 5000 {
		return fmt.Errorf("%w: text length must be 1..5000", ErrValidation)
	}
	return nil
}
