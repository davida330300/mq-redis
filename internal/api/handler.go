package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

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
	if len(req.Payload) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: ErrMissingPayload})
		return
	}
	if len(req.Payload) > h.maxPayloadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: ErrPayloadTooLarge})
		return
	}

	ctx := c.Request.Context()
	jobID, found, err := h.store.GetJobIDByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		if errors.Is(err, ErrStoreUnavailable) {
			h.failOpen(c, req)
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrStore})
		return
	}
	if found {
		c.JSON(http.StatusCreated, JobResponse{JobID: jobID, Status: string(state.Queued)})
		return
	}

	jobID, err = newJobID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrIDGeneration})
		return
	}
	if err := h.store.CreateJob(ctx, req.IdempotencyKey, jobID, req.Payload); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrStore})
		return
	}
	if err := h.producer.Publish(ctx, jobID, req.Payload); err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: ErrPublish})
		return
	}

	c.JSON(http.StatusCreated, JobResponse{JobID: jobID, Status: string(state.Queued)})
}

func (h *Handler) failOpen(c *gin.Context, req JobRequest) {
	jobID, err := newJobID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: ErrIDGeneration})
		return
	}
	if err := h.producer.Publish(c.Request.Context(), jobID, req.Payload); err != nil {
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
