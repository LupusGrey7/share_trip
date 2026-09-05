-- +goose Up
-- +goose StatementBegin

-- event_id: The same UUID that goes to Kafka (idempotency on the consumer).
-- Initially nullable — the table may already contain an old row without an event ID.
ALTER TABLE outbox_event
ADD column IF NOT EXISTS event_id uuid;

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
