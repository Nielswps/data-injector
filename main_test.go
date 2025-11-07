package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type mockRedis struct {
	pingErr   error
	setErrors map[string]error
	pingCount int
	setCount  int
}

func (m *mockRedis) Ping(ctx context.Context) *redis.StatusCmd {
	m.pingCount++
	cmd := redis.NewStatusCmd(ctx)
	if m.pingErr != nil {
		cmd.SetErr(m.pingErr)
	}
	return cmd
}

func (m *mockRedis) JSONSet(ctx context.Context, key string, path string, value interface{}) *redis.StatusCmd {
	m.setCount++
	cmd := redis.NewStatusCmd(ctx)
	if err, ok := m.setErrors[key]; ok {
		cmd.SetErr(err)
	}
	return cmd
}

func (m *mockRedis) Close() error { return nil }

func TestRun_Success(t *testing.T) {
	tmpFile := "testdata.json"
	jsonData := `[{"key":"user1","value":{"name":"Alice"}}]`
	os.WriteFile(tmpFile, []byte(jsonData), 0644)
	defer os.Remove(tmpFile)

	mock := &mockRedis{}
	cfg := Config{RedisAddress: "localhost:6379", DataPath: tmpFile, PingMaxAttempts: 3, PingAttemptTimeout: 2 * time.Millisecond}

	err := run(context.Background(), cfg, mock)
	assert.NoError(t, err)
	assert.Equal(t, 1, mock.pingCount)
	assert.Equal(t, 1, mock.setCount)
}

func TestRun_FailsToConnect(t *testing.T) {
	cfg := Config{RedisAddress: "bad:6379", DataPath: "doesntmatter", PingMaxAttempts: 3, PingAttemptTimeout: 2 * time.Millisecond}
	mock := &mockRedis{pingErr: errors.New("cannot connect")}

	err := run(context.Background(), cfg, mock)
	assert.ErrorContains(t, err, "unable to connect to Redis")
}

func TestRun_InvalidJSON(t *testing.T) {
	tmpFile := "invalid.json"
	os.WriteFile(tmpFile, []byte(`not json`), 0644)
	defer os.Remove(tmpFile)

	mock := &mockRedis{}
	cfg := Config{RedisAddress: "localhost:6379", DataPath: tmpFile, PingMaxAttempts: 3, PingAttemptTimeout: 2 * time.Millisecond}

	err := run(context.Background(), cfg, mock)
	assert.ErrorContains(t, err, "could not unmarshal JSON")
}

func TestRun_MissingKey(t *testing.T) {
	tmpFile := "missingkey.json"
	os.WriteFile(tmpFile, []byte(`[{"value":"test"}]`), 0644)
	defer os.Remove(tmpFile)

	mock := &mockRedis{}
	cfg := Config{RedisAddress: "localhost:6379", DataPath: tmpFile, PingMaxAttempts: 3, PingAttemptTimeout: 2 * time.Millisecond}

	err := run(context.Background(), cfg, mock)
	assert.ErrorContains(t, err, "missing key")
}
