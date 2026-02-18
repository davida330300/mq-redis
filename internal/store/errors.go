package store

import "errors"

var (
	ErrAlreadyExists    = errors.New("idempotency key already exists")
	ErrStoreUnavailable = errors.New("store unavailable")
)
