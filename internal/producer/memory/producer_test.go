package memory

import (
	"context"
	"encoding/json"
	"testing"
)

func TestProducer_Publish(t *testing.T) {
	producer := New()
	payload := json.RawMessage(`{"a":1}`)

	if err := producer.Publish(context.Background(), "job1", payload); err != nil {
		t.Fatalf("Publish error: %v", err)
	}

	published := producer.Published()
	if len(published) != 1 {
		t.Fatalf("published len = %d, want 1", len(published))
	}
	if published[0].JobID != "job1" {
		t.Fatalf("jobID = %q, want %q", published[0].JobID, "job1")
	}
	if string(published[0].Payload) != string(payload) {
		t.Fatalf("payload mismatch")
	}
}
