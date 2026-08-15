package contracts

import "errors"

const (
	ErrServiceUnavailable = "service is not available"
)

// Sentinel-ошибки клиента — «какие отказы бывают» (аналог разных error-классов в Java).
// Позже: ErrTimeout, ErrUnavailable, ErrBadRequest — для errors.Is в handler.
var (
	ErrCheckServiceUnavailable = errors.New("service is not available in the contract")
)
