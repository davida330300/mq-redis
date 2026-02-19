package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mq-redis/internal/idempotency"
	"mq-redis/internal/payload"
	"mq-redis/internal/state"
)

type Handler struct {
	store           Store
	producer        Producer
	maxPayloadBytes int
}

func NewHandler(store Store, producer Producer) *Handler {
	return &Handler{
		store:           store,
		producer:        producer,
		maxPayloadBytes: MaxPayloadBytes,
	}
}

func NewRouter(store Store, producer Producer) *gin.Engine {
	r := gin.New()
	h := NewHandler(store, producer)
	r.POST("/jobs", h.PostJobs)
	return r
}

func (h *Handler) PostJobs(c *gin.Context) {
	var req JobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrInvalidJSON})
		return
	}

	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrMissingIdempotency})
		return
	}
	decision, jobPayload, err := payload.Normalize(payload.Input{
		Inline: req.Payload,
		Ref:    req.PayloadRef,
		Size:   req.PayloadSize,
		Hash:   req.PayloadHash,
	}, h.maxPayloadBytes)
	switch decision {
	case payload.DecisionMissing:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrMissingPayload})
		return
	case payload.DecisionConflict:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrPayloadConflict})
		return
	case payload.DecisionInlineTooLarge:
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: ErrPayloadRefRequired})
		return
	case payload.DecisionRefMetaMissing:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrPayloadRefInvalid})
		return
	case payload.DecisionError:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrPayloadEncoding})
		return
	case payload.DecisionInline, payload.DecisionRef:
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrPayloadEncoding})
		return
	}

	ctx := c.Request.Context()
	jobID, found, err := h.store.GetJobIDByIdempotencyKey(ctx, req.IdempotencyKey)
	switch idempotency.DecideLookup(found, err) {
	case idempotency.LookupFailOpen:
		h.failOpen(c, jobPayload)
		return
	case idempotency.LookupError:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrStore})
		return
	case idempotency.LookupExisting:
		c.JSON(http.StatusCreated, JobResponse{JobID: jobID, Status: string(state.Queued)})
		return
	case idempotency.LookupProceed:
	}

	jobID, err = newJobID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrIDGeneration})
		return
	}
	if err := h.store.CreateJob(ctx, req.IdempotencyKey, jobID, jobPayload); err != nil {
		switch idempotency.DecideCreate(err) {
		case idempotency.CreateAlreadyExists:
			jobID, found, err := h.store.GetJobIDByIdempotencyKey(ctx, req.IdempotencyKey)
			if idempotency.DecideDuplicate(found, err) == idempotency.DuplicateReturnExisting {
				c.JSON(http.StatusCreated, JobResponse{JobID: jobID, Status: string(state.Queued)})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrStore})
			return
		case idempotency.CreateFailOpen:
			h.failOpenWithJobID(c, jobPayload, jobID)
			return
		case idempotency.CreateError:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrStore})
			return
		case idempotency.CreateOK:
		}
	}
	if err := h.producer.Publish(ctx, jobID, jobPayload); err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: ErrPublish})
		return
	}

	c.JSON(http.StatusCreated, JobResponse{JobID: jobID, Status: string(state.Queued)})
}

func (h *Handler) failOpen(c *gin.Context, payload json.RawMessage) {
	jobID, err := newJobID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrIDGeneration})
		return
	}
	h.failOpenWithJobID(c, payload, jobID)
}

func (h *Handler) failOpenWithJobID(c *gin.Context, payload json.RawMessage, jobID string) {
	if err := h.producer.Publish(c.Request.Context(), jobID, payload); err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: ErrPublish})
		return
	}

	c.JSON(http.StatusAccepted, JobResponse{
		JobID:   jobID,
		Status:  string(state.Queued),
		Warning: WarningDedupeDegraded,
	})
}

func newJobID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
