package memory

import (
	"context"
	"encoding/json"
	"testing"
)

func TestStore_CreateAndGet(t *testing.T) {
	store := New()
	payload := json.RawMessage(`{"a":1}`)

	if err := store.CreateJob(context.Background(), "key1", "job1", payload); err != nil {
		t.Fatalf("CreateJob error: %v", err)
	}

	jobID, found, err := store.GetJobIDByIdempotencyKey(context.Background(), "key1")
	if err != nil {
		t.Fatalf("GetJobIDByIdempotencyKey error: %v", err)
	}
	if !found {
		t.Fatalf("expected to find key1")
	}
	if jobID != "job1" {
		t.Fatalf("jobID = %q, want %q", jobID, "job1")
	}
}

func TestStore_CreateDuplicate(t *testing.T) {
	store := New()
	payload := json.RawMessage(`{"a":1}`)

	if err := store.CreateJob(context.Background(), "key1", "job1", payload); err != nil {
		t.Fatalf("CreateJob error: %v", err)
	}
	if err := store.CreateJob(context.Background(), "key1", "job2", payload); err == nil {
		t.Fatalf("expected duplicate create to fail")
	}
}

func TestStore_GetMissing(t *testing.T) {
	store := New()
	jobID, found, err := store.GetJobIDByIdempotencyKey(context.Background(), "missing")
	if err != nil {
		t.Fatalf("GetJobIDByIdempotencyKey error: %v", err)
	}
	if found {
		t.Fatalf("expected missing key to be not found")
	}
	if jobID != "" {
		t.Fatalf("expected empty jobID for missing key")
	}
}
