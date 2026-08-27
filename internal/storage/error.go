package storage

import "errors"

const (
	errQueryByID            = "query trip by ID %s: %w"
	errQueryTripHistoryByID = "query trip_history by id %s: %w"
)

var ErrTripNotFound = errors.New("trip not found")
