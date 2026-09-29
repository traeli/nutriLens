package service

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrForbidden    = errors.New("forbidden")
	ErrUnavailable  = errors.New("service unavailable")
	ErrRateLimited  = errors.New("rate limited")
	ErrDisabled     = errors.New("feature disabled")
)
