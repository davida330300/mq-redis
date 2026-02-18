package rediskeys

import (
	"testing"
	"time"
)

func TestJobKey(t *testing.T) {
	got := JobKey("abc")
	want := "job:abc"
	if got != want {
		t.Fatalf("JobKey() = %q, want %q", got, want)
	}
}

func TestJobDataKey(t *testing.T) {
	got := JobDataKey("xyz")
	want := "job:data:xyz"
	if got != want {
		t.Fatalf("JobDataKey() = %q, want %q", got, want)
	}
}

func TestConstants(t *testing.T) {
	if JobKeyPrefix != "job:" {
		t.Fatalf("JobKeyPrefix = %q, want %q", JobKeyPrefix, "job:")
	}
	if JobDataKeyPrefix != "job:data:" {
		t.Fatalf("JobDataKeyPrefix = %q, want %q", JobDataKeyPrefix, "job:data:")
	}
	if RetryJobsKey != "retry:jobs" {
		t.Fatalf("RetryJobsKey = %q, want %q", RetryJobsKey, "retry:jobs")
	}
	if RetryLockKey != "retry:lock" {
		t.Fatalf("RetryLockKey = %q, want %q", RetryLockKey, "retry:lock")
	}

	if DedupeTTL != 72*time.Hour {
		t.Fatalf("DedupeTTL = %v, want 72h", DedupeTTL)
	}
	if JobStatusTTL != 14*24*time.Hour {
		t.Fatalf("JobStatusTTL = %v, want 14d", JobStatusTTL)
	}
	if JobDataTTL != 14*24*time.Hour {
		t.Fatalf("JobDataTTL = %v, want 14d", JobDataTTL)
	}
	if DLQTTL != 14*24*time.Hour {
		t.Fatalf("DLQTTL = %v, want 14d", DLQTTL)
	}
}
