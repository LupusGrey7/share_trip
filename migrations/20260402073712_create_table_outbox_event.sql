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

comment
    on table outbox_event is 'сохранение события о состояния поездки';
comment
    on column outbox_event.event_name is 'технический статус обработки';
comment
    on column outbox_event.aggregate_id is 'идентификатор агрегата';
comment
    on column outbox_event.payload is 'полезную нагрузку события';
comment
    on column outbox_event.created_at is 'время создания';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table IF EXISTS outbox_event;
-- +goose StatementEnd
