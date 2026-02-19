//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	segkafka "github.com/segmentio/kafka-go"
)

const (
	composeFile      = "docker-compose.yml"
	configPath       = "config/config.e2e.yaml"
	apiAddr          = "http://127.0.0.1:18080"
	jobsTopic        = "jobs-e2e"
	dlqTopic         = "jobs-dlq-e2e"
	redisAddr        = "localhost:6379"
	kafkaBroker      = "localhost:9094"
	postgresDSN      = "postgres://mq:mq@localhost:5432/mq?sslmode=disable"
	connectTimeout   = 2 * time.Second
	testTimeout      = 60 * time.Second
	apiStartTimeout  = 20 * time.Second
	kafkaWaitTimeout = 45 * time.Second
	topicTimeout     = 30 * time.Second
)

var (
	apiCmd     *exec.Cmd
	apiBin     string
	apiTempDir string
)

func TestMain(m *testing.M) {
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Fprintf(os.Stderr, "docker not found: %v\n", err)
		os.Exit(1)
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "repo root: %v\n", err)
		os.Exit(1)
	}

	apiBin, apiTempDir, err = buildAPI(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build api: %v\n", err)
		os.Exit(1)
	}

	if err := dockerComposeUp(repoRoot); err != nil {
		fmt.Fprintf(os.Stderr, "docker compose up: %v\n", err)
		os.Exit(1)
	}

	keepDocker := os.Getenv("E2E_KEEP_DOCKER") == "1"

	ctx, cancel := context.WithTimeout(context.Background(), kafkaWaitTimeout)
	if err := waitForKafka(ctx); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "kafka not ready: %v\n", err)
		os.Exit(1)
	}
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), topicTimeout)
	if err := createTopicWithRetry(ctx, jobsTopic); err != nil {
		fmt.Fprintf(os.Stderr, "create topic: %v\n", err)
		os.Exit(1)
	}
	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), topicTimeout)
	if err := createTopicWithRetry(ctx, dlqTopic); err != nil {
		fmt.Fprintf(os.Stderr, "create dlq topic: %v\n", err)
		os.Exit(1)
	}
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), connectTimeout)
	if err := waitForRedis(ctx); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "redis not ready: %v\n", err)
		os.Exit(1)
	}
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), connectTimeout)
	if err := waitForPostgres(ctx); err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "postgres not ready: %v\n", err)
		os.Exit(1)
	}
	cancel()

	apiCmd, err = startAPI(repoRoot, apiBin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start api: %v\n", err)
		os.Exit(1)
	}

	if err := waitForHealth(apiStartTimeout); err != nil {
		fmt.Fprintf(os.Stderr, "api not ready: %v\n", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	stopAPI()
	if apiTempDir != "" {
		_ = os.RemoveAll(apiTempDir)
	}
	if !keepDocker {
		if err := dockerComposeDown(repoRoot); err != nil {
			fmt.Fprintf(os.Stderr, "docker compose down: %v\n", err)
		}
	}
	os.Exit(exitCode)
}

func TestE2ESmoke(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	idempotencyKey := fmt.Sprintf("e2e-%d", time.Now().UnixNano())
	payload := []byte(`{"hello":"world"}`)

	jobID, err := postJob(ctx, idempotencyKey, payload)
	if err != nil {
		t.Fatalf("post job: %v", err)
	}

	if err := assertRedisKeys(ctx, idempotencyKey, jobID, payload); err != nil {
		t.Fatalf("redis check: %v", err)
	}

	if err := assertKafkaMessage(ctx, jobID, payload); err != nil {
		t.Fatalf("kafka check: %v", err)
	}

	if err := assertPostgres(ctx); err != nil {
		t.Fatalf("postgres check: %v", err)
	}
}

func postJob(ctx context.Context, key string, payload []byte) (string, error) {
	body := map[string]any{
		"idempotency_key": key,
		"payload":         json.RawMessage(payload),
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiAddr+"/jobs", bytes.NewReader(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status: %d body=%s", resp.StatusCode, string(data))
	}

	var out struct {
		JobID string `json:"job_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.JobID == "" {
		return "", fmt.Errorf("missing job_id")
	}
	return out.JobID, nil
}

func assertRedisKeys(ctx context.Context, key, jobID string, payload []byte) error {
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer client.Close()

	val, err := client.Get(ctx, "idem:"+key).Result()
	if err != nil {
		return fmt.Errorf("idempotency key: %w", err)
	}
	if val != jobID {
		return fmt.Errorf("idempotency value = %q", val)
	}
	status, err := client.Get(ctx, "job:"+jobID).Result()
	if err != nil {
		return fmt.Errorf("job status: %w", err)
	}
	if status != "queued" {
		return fmt.Errorf("job status = %q", status)
	}
	data, err := client.Get(ctx, "job:data:"+jobID).Result()
	if err != nil {
		return fmt.Errorf("job data: %w", err)
	}
	if strings.TrimSpace(data) != strings.TrimSpace(string(payload)) {
		return fmt.Errorf("job data mismatch")
	}
	return nil
}

func assertKafkaMessage(ctx context.Context, jobID string, payload []byte) error {
	reader := segkafka.NewReader(segkafka.ReaderConfig{
		Brokers:   []string{kafkaBroker},
		Topic:     jobsTopic,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(10 * time.Second)
	}
	for time.Now().Before(deadline) {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				break
			}
			return err
		}
		if string(msg.Key) == jobID {
			if string(msg.Value) != string(payload) {
				return fmt.Errorf("payload mismatch")
			}
			return nil
		}
	}
	return fmt.Errorf("no kafka message for job")
}

func assertPostgres(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, postgresDSN)
	if err != nil {
		return err
	}
	defer pool.Close()
	return pool.Ping(ctx)
}

func waitForHealth(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(apiAddr + "/healthz")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("healthz timeout")
}

func waitForRedis(ctx context.Context) error {
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer client.Close()
	return client.Ping(ctx).Err()
}

func waitForPostgres(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, postgresDSN)
	if err != nil {
		return err
	}
	defer pool.Close()
	return pool.Ping(ctx)
}

func waitForKafka(ctx context.Context) error {
	for {
		conn, err := segkafka.DialContext(ctx, "tcp", kafkaBroker)
		if err == nil {
			conn.Close()
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func createTopicWithRetry(ctx context.Context, topic string) error {
	for {
		if err := createTopic(topic); err == nil {
			return nil
		} else if strings.Contains(err.Error(), "Topic with this name already exists") {
			return nil
		} else if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func createTopic(topic string) error {
	conn, err := segkafka.Dial("tcp", kafkaBroker)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	controllerConn, err := segkafka.Dial("tcp", netAddress(controller.Host, controller.Port))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	cfg := segkafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}
	err = controllerConn.CreateTopics(cfg)
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "Topic with this name already exists") {
		return nil
	}
	return err
}

func dockerComposeUp(dir string) error {
	cmd := exec.Command("docker", "compose", "-f", filepath.Join(dir, composeFile), "up", "-d")
	if os.Getenv("E2E_VERBOSE") == "1" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
	return cmd.Run()
}

func dockerComposeDown(dir string) error {
	cmd := exec.Command("docker", "compose", "-f", filepath.Join(dir, composeFile), "down")
	if os.Getenv("E2E_VERBOSE") == "1" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
	return cmd.Run()
}

func buildAPI(dir string) (string, string, error) {
	tmpDir, err := os.MkdirTemp("", "mq-redis-e2e-*")
	if err != nil {
		return "", "", err
	}
	binPath := filepath.Join(tmpDir, "api-e2e")
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/api")
	cmd.Dir = dir
	if os.Getenv("E2E_VERBOSE") == "1" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", "", err
	}
	return binPath, tmpDir, nil
}

func startAPI(dir, binPath string) (*exec.Cmd, error) {
	cmd := exec.Command(binPath)
	cmd.Dir = dir
	if os.Getenv("E2E_VERBOSE") == "1" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
	cmd.Env = append(os.Environ(), "CONFIG_PATH="+filepath.Join(dir, configPath))
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func stopAPI() {
	if apiCmd == nil || apiCmd.Process == nil {
		return
	}
	_ = apiCmd.Process.Signal(os.Interrupt)
	done := make(chan struct{})
	go func() {
		_, _ = apiCmd.Process.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = apiCmd.Process.Kill()
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root not found")
		}
		dir = parent
	}
}

func netAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
