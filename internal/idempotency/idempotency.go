package idempotency

import (
	"errors"

	"mq-redis/internal/store"
)

type LookupDecision int

const (
	LookupExisting LookupDecision = iota
	LookupProceed
	LookupFailOpen
	LookupError
)

func DecideLookup(found bool, err error) LookupDecision {
	if err != nil {
		if errors.Is(err, store.ErrStoreUnavailable) {
			return LookupFailOpen
		}
		return LookupError
	}
	if found {
		return LookupExisting
	}
	return LookupProceed
}

type CreateDecision int

const (
	CreateOK CreateDecision = iota
	CreateAlreadyExists
	CreateFailOpen
	CreateError
)

func DecideCreate(err error) CreateDecision {
	if err == nil {
		return CreateOK
	}
	if errors.Is(err, store.ErrAlreadyExists) {
		return CreateAlreadyExists
	}
	if errors.Is(err, store.ErrStoreUnavailable) {
		return CreateFailOpen
	}
	return CreateError
}

type DuplicateDecision int

const (
	DuplicateReturnExisting DuplicateDecision = iota
	DuplicateError
)

func DecideDuplicate(found bool, err error) DuplicateDecision {
	if err != nil || !found {
		return DuplicateError
	}
	return DuplicateReturnExisting
}
