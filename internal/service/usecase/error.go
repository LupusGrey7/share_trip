package usecase

import "errors"

var (
	ErrTripNotFound = errors.New("trip not found")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrBadRequest   = errors.New("bad request: invalid entity status, expected is %s, got %s")
)
