package chat

import "errors"

// ErrNotFound используем, когда сущность не найдена (например чат по id).
var ErrNotFound = errors.New("not found")

// ErrValidation используем для ошибок валидации входных данных (title/text/limit).
var ErrValidation = errors.New("validation error")
