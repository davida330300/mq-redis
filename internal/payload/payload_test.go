package payload

import (
	"encoding/json"
	"testing"
)

func TestNormalizeInline(t *testing.T) {
	input := Input{Inline: json.RawMessage(`{"a":1}`)}
	decision, payload, err := Normalize(input, 1024)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if decision != DecisionInline {
		t.Fatalf("decision = %v, want %v", decision, DecisionInline)
	}
	if string(payload) != string(input.Inline) {
		t.Fatalf("payload mismatch: got %s", payload)
	}
}

func TestNormalizeInlineTooLarge(t *testing.T) {
	input := Input{Inline: json.RawMessage(`{"a":1}`)}
	decision, payload, err := Normalize(input, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if decision != DecisionInlineTooLarge {
		t.Fatalf("decision = %v, want %v", decision, DecisionInlineTooLarge)
	}
	if payload != nil {
		t.Fatalf("expected nil payload")
	}
}

func TestNormalizeRef(t *testing.T) {
	input := Input{Ref: " s3://bucket/key ", Size: 123, Hash: " abc "}
	decision, payload, err := Normalize(input, 1024)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if decision != DecisionRef {
		t.Fatalf("decision = %v, want %v", decision, DecisionRef)
	}
	var got RefPayload
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if got.Ref != "s3://bucket/key" || got.Size != 123 || got.Hash != "abc" {
		t.Fatalf("unexpected ref payload: %+v", got)
	}
}

func TestNormalizeRefMissingMeta(t *testing.T) {
	input := Input{Ref: "s3://bucket/key", Size: 0, Hash: ""}
	decision, payload, err := Normalize(input, 1024)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if decision != DecisionRefMetaMissing {
		t.Fatalf("decision = %v, want %v", decision, DecisionRefMetaMissing)
	}
	if payload != nil {
		t.Fatalf("expected nil payload")
	}
}

func TestNormalizeMissing(t *testing.T) {
	input := Input{}
	decision, payload, err := Normalize(input, 1024)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if decision != DecisionMissing {
		t.Fatalf("decision = %v, want %v", decision, DecisionMissing)
	}
	if payload != nil {
		t.Fatalf("expected nil payload")
	}
}

func TestNormalizeConflict(t *testing.T) {
	input := Input{Inline: json.RawMessage(`{"a":1}`), Ref: "s3://bucket/key"}
	decision, payload, err := Normalize(input, 1024)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if decision != DecisionConflict {
		t.Fatalf("decision = %v, want %v", decision, DecisionConflict)
	}
	if payload != nil {
		t.Fatalf("expected nil payload")
	}
}
