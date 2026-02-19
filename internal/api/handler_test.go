package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	storeerr "mq-redis/internal/store"
)

type fakeStore struct {
	getKey        string
	getJobID      string
	getFound      bool
	getErr        error
	getCalls      int
	getResults    []getResult
	createCalled  bool
	createKey     string
	createJobID   string
	createPayload json.RawMessage
	createErr     error
}

type getResult struct {
	jobID string
	found bool
	err   error
}

func (s *fakeStore) GetJobIDByIdempotencyKey(ctx context.Context, key string) (string, bool, error) {
	s.getKey = key
	if len(s.getResults) > 0 {
		idx := s.getCalls
		s.getCalls++
		if idx < len(s.getResults) {
			res := s.getResults[idx]
			return res.jobID, res.found, res.err
		}
	}
	return s.getJobID, s.getFound, s.getErr
}

func (s *fakeStore) CreateJob(ctx context.Context, key, jobID string, payload json.RawMessage) error {
	s.createCalled = true
	s.createKey = key
	s.createJobID = jobID
	s.createPayload = payload
	return s.createErr
}

type fakeProducer struct {
	publishCalled  bool
	publishJobID   string
	publishPayload json.RawMessage
	publishErr     error
}

func (p *fakeProducer) Publish(ctx context.Context, jobID string, payload json.RawMessage) error {
	p.publishCalled = true
	p.publishJobID = jobID
	p.publishPayload = payload
	return p.publishErr
}

func TestPostJobs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeStore{}
	producer := &fakeProducer{}
	r := NewRouter(store, producer)

	body := []byte(`{"idempotency_key":"k1","payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if !store.createCalled {
		t.Fatalf("expected CreateJob to be called")
	}
	if !producer.publishCalled {
		t.Fatalf("expected Publish to be called")
	}
	if store.createJobID == "" {
		t.Fatalf("expected job id to be set")
	}
	if store.createJobID != producer.publishJobID {
		t.Fatalf("job id mismatch between store and producer")
	}
}

func TestPostJobs_Duplicate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeStore{getFound: true, getJobID: "job-123"}
	producer := &fakeProducer{}
	r := NewRouter(store, producer)

	body := []byte(`{"idempotency_key":"k1","payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if store.createCalled {
		t.Fatalf("did not expect CreateJob to be called")
	}
	if producer.publishCalled {
		t.Fatalf("did not expect Publish to be called")
	}
	if !strings.Contains(w.Body.String(), "job-123") {
		t.Fatalf("expected response to contain original job id")
	}
}

func TestPostJobs_MissingIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(&fakeStore{}, &fakeProducer{})

	body := []byte(`{"payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPostJobs_PayloadTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(&fakeStore{}, &fakeProducer{})

	big := strings.Repeat("a", MaxPayloadBytes+1)
	body := []byte(`{"idempotency_key":"k1","payload":"` + big + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestPostJobs_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(&fakeStore{}, &fakeProducer{})

	body := []byte(`{"idempotency_key":"k1","payload":`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPostJobs_FailOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeStore{getErr: storeerr.ErrStoreUnavailable}
	producer := &fakeProducer{}
	r := NewRouter(store, producer)

	body := []byte(`{"idempotency_key":"k1","payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusAccepted)
	}
	if producer.publishCalled != true {
		t.Fatalf("expected Publish to be called")
	}
	if store.createCalled {
		t.Fatalf("did not expect CreateJob to be called")
	}
	if !strings.Contains(w.Body.String(), WarningDedupeDegraded) {
		t.Fatalf("expected warning in response")
	}
}

func TestPostJobs_CreateFailOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeStore{createErr: storeerr.ErrStoreUnavailable}
	producer := &fakeProducer{}
	r := NewRouter(store, producer)

	body := []byte(`{"idempotency_key":"k1","payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusAccepted)
	}
	if producer.publishCalled != true {
		t.Fatalf("expected Publish to be called")
	}
	if store.createJobID == "" {
		t.Fatalf("expected job id to be generated")
	}
	if producer.publishJobID != store.createJobID {
		t.Fatalf("expected publish to use the generated job id")
	}
	if !strings.Contains(w.Body.String(), WarningDedupeDegraded) {
		t.Fatalf("expected warning in response")
	}
}

func TestPostJobs_DuplicateCreateRace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeStore{
		getResults: []getResult{
			{jobID: "", found: false, err: nil},
			{jobID: "job-123", found: true, err: nil},
		},
		createErr: storeerr.ErrAlreadyExists,
	}
	producer := &fakeProducer{}
	r := NewRouter(store, producer)

	body := []byte(`{"idempotency_key":"k1","payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if !strings.Contains(w.Body.String(), "job-123") {
		t.Fatalf("expected response to contain original job id")
	}
}

func TestPostJobs_PublishFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeStore{}
	producer := &fakeProducer{publishErr: errors.New("publish failed")}
	r := NewRouter(store, producer)

	body := []byte(`{"idempotency_key":"k1","payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestPostJobs_StoreFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeStore{getErr: errors.New("store failure")}
	producer := &fakeProducer{}
	r := NewRouter(store, producer)

	body := []byte(`{"idempotency_key":"k1","payload":{"a":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
