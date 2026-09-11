-- +goose Up
-- +goose StatementBegin

-- Предыдущая 20260905162931 могла примениться как goose-шаблон (SELECT 'up SQL query')
-- без реального ALTER — колонки event_id в БД нет, а version уже Applied.
-- Эта миграция добавляет event_id идемпотентно.

ALTER TABLE outbox_event
    ADD COLUMN IF NOT EXISTS event_id uuid;

UPDATE outbox_event
SET event_id = gen_random_uuid()
WHERE event_id IS NULL;

ALTER TABLE outbox_event
    ALTER COLUMN event_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_outbox_event_event_id ON outbox_event (event_id);

COMMENT ON COLUMN outbox_event.event_id IS 'stable event UUID shared with Kafka message';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_outbox_event_event_id;
ALTER TABLE outbox_event DROP COLUMN IF EXISTS event_id;

-- +goose StatementEnd
