package api

import "encoding/json"

const MaxPayloadBytes = 256 * 1024

const WarningDedupeDegraded = "dedupe_degraded"

const (
	ErrInvalidJSON        = "invalid_json"
	ErrMissingIdempotency = "missing_idempotency_key"
	ErrMissingPayload     = "missing_payload"
	ErrPayloadTooLarge    = "payload_too_large"
	ErrPayloadConflict    = "payload_conflict"
	ErrPayloadRefRequired = "payload_ref_required"
	ErrPayloadRefInvalid  = "payload_ref_invalid"
	ErrPayloadEncoding    = "payload_encoding_failed"
	ErrStore              = "store_error"
	ErrPublish            = "publish_failed"
	ErrIDGeneration       = "id_generation_failed"
)

type JobRequest struct {
	IdempotencyKey string          `json:"idempotency_key"`
	Payload        json.RawMessage `json:"payload"`
	PayloadRef     string          `json:"payload_ref,omitempty"`
	PayloadSize    int64           `json:"payload_size,omitempty"`
	PayloadHash    string          `json:"payload_hash,omitempty"`
}

type JobResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status,omitempty"`
	Warning string `json:"warning,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
