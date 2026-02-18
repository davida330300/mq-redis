package state

import "testing"

func TestCanTransition_AllowsExpected(t *testing.T) {
	cases := []struct {
		from State
		to   State
	}{
		{Queued, Processing},
		{Processing, Done},
		{Processing, Retrying},
		{Processing, DLQ},
		{Retrying, Queued},
		{SagaRunning, SagaStepFailed},
		{SagaStepFailed, Retrying},
		{SagaCompensating, SagaCompensated},
		{SagaCompensated, DLQ},
	}

	for _, tc := range cases {
		if !CanTransition(tc.from, tc.to) {
			t.Fatalf("expected transition %q -> %q to be allowed", tc.from, tc.to)
		}
	}
}

func TestCanTransition_BlocksUnexpected(t *testing.T) {
	cases := []struct {
		from State
		to   State
	}{
		{Queued, Done},
		{Done, Processing},
		{DLQ, Processing},
		{Retrying, Processing},
		{SagaRunning, DLQ},
	}

	for _, tc := range cases {
		if CanTransition(tc.from, tc.to) {
			t.Fatalf("expected transition %q -> %q to be blocked", tc.from, tc.to)
		}
	}
}

func TestIsTerminal(t *testing.T) {
	if !IsTerminal(Done) {
		t.Fatalf("expected Done to be terminal")
	}
	if !IsTerminal(DLQ) {
		t.Fatalf("expected DLQ to be terminal")
	}
	if IsTerminal(Queued) {
		t.Fatalf("expected Queued to be non-terminal")
	}
	if IsTerminal(SagaCompensated) {
		t.Fatalf("expected SagaCompensated to be non-terminal")
	}
}

func TestAllStates(t *testing.T) {
	got := AllStates()
	if len(got) != len(allStates) {
		t.Fatalf("AllStates length = %d, want %d", len(got), len(allStates))
	}

	seen := map[State]bool{}
	for _, s := range got {
		if seen[s] {
			t.Fatalf("duplicate state %q", s)
		}
		seen[s] = true
	}

	for _, s := range allStates {
		if !seen[s] {
			t.Fatalf("missing state %q", s)
		}
	}
}
