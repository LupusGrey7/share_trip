-- +goose Up
-- +goose StatementBegin

-- legacy table from early drafts (if present on local DB)
DROP TABLE IF EXISTS outbox_event;

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_events_pending
    ON outbox_events (created_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate_id
    ON outbox_events (aggregate_id);
CREATE INDEX IF NOT EXISTS idx_outbox_events_event_type
    ON outbox_events (event_type);

COMMENT ON TABLE outbox_events IS 'transactional outbox: trip events pending Kafka delivery';
COMMENT ON COLUMN outbox_events.id IS 'event identifier (idempotency key for consumers)';
COMMENT ON COLUMN outbox_events.aggregate_type IS 'aggregate type, e.g. trip';
COMMENT ON COLUMN outbox_events.aggregate_id IS 'aggregate id (trip_id)';
COMMENT ON COLUMN outbox_events.event_type IS 'event type, e.g. trip_published';
COMMENT ON COLUMN outbox_events.payload IS 'JSON body for Kafka envelope.payload';
COMMENT ON COLUMN outbox_events.status IS 'pending | sent | failed';
COMMENT ON COLUMN outbox_events.attempts IS 'number of produce attempts';
COMMENT ON COLUMN outbox_events.last_error IS 'last produce error';
COMMENT ON COLUMN outbox_events.created_at IS 'when the outbox row was created';
COMMENT ON COLUMN outbox_events.sent_at IS 'when successfully produced to Kafka';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_outbox_events_pending;
DROP INDEX IF EXISTS idx_outbox_events_aggregate_id;
DROP INDEX IF EXISTS idx_outbox_events_event_type;

DROP TABLE IF EXISTS outbox_events;

-- +goose StatementEnd
