-- +goose Up
-- +goose StatementBegin
create table IF NOT EXISTS outbox_event
(
    id           bigserial primary key,
    event_name   text        not null,
    aggregate_id UUID        NOT NULL,
    payload      jsonb       not null, -- JSON с trip_id.
    created_at   timestamptz not null default now()
);

COMMENT
    ON TABLE outbox_event IS 'сохранение события о состояния поездки';
COMMENT
    ON COLUMN outbox_event.event_name IS 'технический статус обработки';
COMMENT
    ON COLUMN outbox_event.aggregate_id IS 'идентификатор агрегата';
COMMENT
    ON COLUMN outbox_event.payload IS 'полезную нагрузку события';
COMMENT
    ON COLUMN outbox_event.created_at IS 'время создания';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox_event;
-- +goose StatementEnd
