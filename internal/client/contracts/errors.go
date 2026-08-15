package contracts

import "errors"

// Sentinel-ошибки исходящего клиента Contract Service.
// HandleError мапит их через errors.Is → 502/503/504 (fail closed).
// Не путать с CheckResult{Allowed:false} — это бизнес-deny без error.
var (
	ErrTimeout     = errors.New("contract service timeout")
	ErrUnavailable = errors.New("contract service unavailable")
	ErrBadRequest  = errors.New("contract service bad request")
	ErrForbidden   = errors.New("contract service forbidden")
)
