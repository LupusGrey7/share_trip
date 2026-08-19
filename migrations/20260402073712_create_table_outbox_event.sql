-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS outbox_event
(
    id bigserial PRIMARY KEY,
    event_id uuid NOT NULL,
    event_name text NOT NULL,
    aggregate_id uuid NOT NULL,
    payload jsonb NOT NULL, -- JSON с trip_id.
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_outbox_event_event_name ON outbox_event (event_name);
CREATE INDEX idx_outbox_event_event_id ON outbox_event (event_id);
CREATE INDEX idx_outbox_event_created_at ON outbox_event (created_at);
CREATE INDEX idx_outbox_event_aggregate_id ON outbox_event (aggregate_id);

COMMENT ON TABLE outbox_event IS 'saving the event of the trip status';
COMMENT ON COLUMN outbox_event.event_name IS 'technical status of processing';
COMMENT ON COLUMN outbox_event.event_id IS 'event identifier (uuid)';
COMMENT ON COLUMN outbox_event.aggregate_id IS 'aggregate identifier';
COMMENT ON COLUMN outbox_event.payload IS 'useful payload of the event';
COMMENT ON COLUMN outbox_event.created_at IS 'creation time';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_outbox_event_event_name;
DROP INDEX IF EXISTS idx_outbox_event_event_id;
DROP INDEX IF EXISTS idx_outbox_event_created_at;
DROP INDEX IF EXISTS idx_outbox_event_aggregate_id;

DROP TABLE IF EXISTS outbox_event;
-- +goose StatementEnd
