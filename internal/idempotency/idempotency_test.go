package idempotency

import (
	"errors"
	"testing"

	"mq-redis/internal/store"
)

func TestDecideLookup(t *testing.T) {
	cases := []struct {
		name  string
		found bool
		err   error
		want  LookupDecision
	}{
		{name: "found", found: true, want: LookupExisting},
		{name: "not-found", found: false, want: LookupProceed},
		{name: "store-unavailable", err: store.ErrStoreUnavailable, want: LookupFailOpen},
		{name: "other-error", err: errors.New("boom"), want: LookupError},
	}

	for _, tc := range cases {
		if got := DecideLookup(tc.found, tc.err); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestDecideCreate(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want CreateDecision
	}{
		{name: "ok", want: CreateOK},
		{name: "already-exists", err: store.ErrAlreadyExists, want: CreateAlreadyExists},
		{name: "store-unavailable", err: store.ErrStoreUnavailable, want: CreateFailOpen},
		{name: "other-error", err: errors.New("boom"), want: CreateError},
	}

	for _, tc := range cases {
		if got := DecideCreate(tc.err); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestDecideDuplicate(t *testing.T) {
	cases := []struct {
		name  string
		found bool
		err   error
		want  DuplicateDecision
	}{
		{name: "found", found: true, want: DuplicateReturnExisting},
		{name: "not-found", found: false, want: DuplicateError},
		{name: "error", found: true, err: errors.New("boom"), want: DuplicateError},
	}

	for _, tc := range cases {
		if got := DecideDuplicate(tc.found, tc.err); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}
