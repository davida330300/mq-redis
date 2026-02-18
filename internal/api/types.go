package api

import "encoding/json"

const MaxPayloadBytes = 256 * 1024

const WarningDedupeDegraded = "dedupe_degraded"

const (
	ErrInvalidJSON        = "invalid_json"
	ErrMissingIdempotency = "missing_idempotency_key"
	ErrMissingPayload     = "missing_payload"
	ErrPayloadTooLarge    = "payload_too_large"
	ErrStore              = "store_error"
	ErrPublish            = "publish_failed"
	ErrIDGeneration       = "id_generation_failed"
)

type JobRequest struct {
	IdempotencyKey string          `json:"idempotency_key"`
	Payload        json.RawMessage `json:"payload"`
}

type JobResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status,omitempty"`
	Warning string `json:"warning,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
