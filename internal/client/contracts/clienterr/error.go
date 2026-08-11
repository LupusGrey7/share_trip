package clienterr

import "errors"

const (
	ErrServiceUnavailable = "service is not available"
)

var (
	ErrCheckServiceUnavailable = errors.New("service is not available in the contract")
)
