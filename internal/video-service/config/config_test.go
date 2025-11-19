package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	validConfig := `
server:
  port: 50052
  read_timeout: 30s
  write_timeout: 30s

postgresql:
  host: postgres-test
  port: 5433
  user: testuser
  password: testpass
  name: video_test

redis:
  host: redis-test
  port: 6380
  password: testpass
  db: 1
  cache_ttl: 2h

kafka:
  brokers:
    - kafka1:9092
  publish_topic: video.metadata.created
  consume_topics:
    - video.uploaded
    - video.transcoded
  group_id: video-test-group

logging:
  level: info
  format: json
  output: stdout
`
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := Loader()

	assert.Equal(t, 50052, cfg.Server.Port)
	assert.Equal(t, "postgres-test", cfg.PostgreSQL.Host)
	assert.Equal(t, 5433, cfg.PostgreSQL.Port)
	assert.Equal(t, "testuser", cfg.PostgreSQL.User)
	assert.Equal(t, "redis-test", cfg.Redis.Host)
	assert.Equal(t, 2*time.Hour, cfg.Redis.CacheTTL)
	assert.Equal(t, []string{"kafka1:9092"}, cfg.Kafka.Brokers)
	assert.Equal(t, "video.metadata.created", cfg.Kafka.PublishTopic)
	assert.Equal(t, "info", cfg.Logging.Level)
}

func TestLoader_DefaultValues(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	minimalConfig := `
postgresql:
  host: localhost
  port: 5432
  user: testuser
  password: testpass
  name: video_test

redis:
  host: localhost
  port: 6379

kafka:
  brokers:
    - localhost:9092
`
	err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := Loader()

	assert.Equal(t, 50052, cfg.Server.Port)
	assert.Equal(t, 1*time.Hour, cfg.Redis.CacheTTL)
	assert.Equal(t, "video.metadata.created", cfg.Kafka.PublishTopic)
	assert.Contains(t, cfg.Kafka.ConsumeTopics, "video.uploaded")
	assert.Equal(t, "video-service-group", cfg.Kafka.GroupID)
}

// Note: Tests for missing required fields, invalid YAML, and non-existent config files
// are skipped because Loader() calls log.Fatalf which does os.Exit, not panic.
// These scenarios would terminate the test process, making them unsuitable for unit testing.
// In production, these errors are correctly handled by terminating the service.
