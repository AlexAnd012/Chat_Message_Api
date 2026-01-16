-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS chats (
id  BIGSERIAL PRIMARY KEY,
title   VARCHAR(200) NOT NULL,
created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messages (
id         BIGSERIAL PRIMARY KEY,
chat_id    BIGINT      NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
text       VARCHAR(5000) NOT NULL,
created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- вспомогательный индекс для запроса последних N сообщений в чате
CREATE INDEX IF NOT EXISTS idx_messages_desc
    ON messages (chat_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_messages_desc;

DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;

-- +goose StatementEnd
