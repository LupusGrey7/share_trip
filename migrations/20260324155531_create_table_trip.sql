-- +goose Up
-- +goose StatementBegin

-- Dictionary instead of PostgreSQL ENUM: add statuses via INSERT, no ALTER TYPE.
CREATE TABLE IF NOT EXISTS trip_status (
    id SMALLSERIAL PRIMARY KEY,
    name VARCHAR(32) NOT NULL UNIQUE -- unique status name
);

-- Fixed ids so DEFAULT and seeds stay stable across reinstalls.
INSERT INTO trip_status (id, name) VALUES
(1, 'draft'),
(2, 'published'),
(3, 'started'),
(4, 'cancelled'),
(5, 'completed');

SELECT setval(
    pg_get_serial_sequence('trip_status', 'id'),
    (SELECT max(id) FROM trip_status)
);

CREATE TABLE IF NOT EXISTS trips
(
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    from_point TEXT NOT NULL,
    to_point TEXT NOT NULL,
    departure_time TIMESTAMPTZ NOT NULL,
    seats INT NOT NULL CHECK (seats > 0),
    trip_status_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_trip_status FOREIGN KEY (trip_status_id)
    REFERENCES trip_status(id) ON DELETE RESTRICT --in case of deleting the status, the trip will not be deleted
);

CREATE TABLE IF NOT EXISTS trip_history
(
    id UUID PRIMARY KEY,
    trip_id UUID NOT NULL,
    from_status_id INT,
    to_status_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_trip_history_trip FOREIGN KEY (trip_id)
    REFERENCES trips(id) ON DELETE CASCADE,
    CONSTRAINT fk_trip_history_from_status FOREIGN KEY (from_status_id)
    REFERENCES trip_status(id) ON DELETE RESTRICT,
    CONSTRAINT fk_trip_history_to_status FOREIGN KEY (to_status_id)
    REFERENCES trip_status(id) ON DELETE RESTRICT
);

CREATE INDEX idx_trips_driver_id ON trips (driver_id);
CREATE INDEX idx_trips_trip_status_id ON trips (trip_status_id);
CREATE INDEX idx_trips_departure_time ON trips (departure_time);
CREATE INDEX idx_trips_seats ON trips (seats);
CREATE INDEX idx_trips_created_at ON trips (created_at);
CREATE INDEX idx_trip_history_trip_id ON trip_history (trip_id);
CREATE INDEX idx_trip_history_from_status_id ON trip_history (from_status_id);
CREATE INDEX idx_trip_history_to_status_id ON trip_history (to_status_id);
CREATE INDEX idx_trip_history_created_at ON trip_history (created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_trips_driver_id;
DROP INDEX IF EXISTS idx_trips_trip_status_id;
DROP INDEX IF EXISTS idx_trips_departure_time;
DROP INDEX IF EXISTS idx_trips_seats;
DROP INDEX IF EXISTS idx_trips_created_at;
DROP INDEX IF EXISTS idx_trip_history_trip_id;
DROP INDEX IF EXISTS idx_trip_history_from_status_id;
DROP INDEX IF EXISTS idx_trip_history_to_status_id;
DROP INDEX IF EXISTS idx_trip_history_created_at;

DROP TABLE IF EXISTS trip_history;
DROP TABLE IF EXISTS trips;
DROP TABLE IF EXISTS trip_status;

-- +goose StatementEnd
