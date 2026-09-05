package contracts

import "errors"

// Sentinel-errors are outgoing client errors from Contract Service.
// HandleError maps them through errors.Is → 502/503/504 (fail closed).
// Don't confuse with CheckResult{Allowed:false} — this is business deny without error.
var (
	ErrTimeout     = errors.New("contract service timeout")     // 504
	ErrUnavailable = errors.New("contract service unavailable") // 502/503/504
	ErrBadRequest  = errors.New("contract service bad request") // 400
	ErrForbidden   = errors.New("contract service forbidden")   // 403
)
