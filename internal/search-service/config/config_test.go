package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	validConfig := `
server:
  port: 50056

elasticsearch:
  addresses:
    - http://es1:9200
    - http://es2:9200
  username: elastic
  password: testpass
  index: videos_test
  shards: 5
  replicas: 2

kafka:
  brokers:
    - kafka1:9092
  consume_topics:
    - video.metadata.created
    - video.metadata.updated
    - video.deleted
  group_id: search-test-group

logging:
  level: info
`
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := Loader()

	assert.Equal(t, 50056, cfg.Server.Port)
	assert.Equal(t, []string{"http://es1:9200", "http://es2:9200"}, cfg.Elasticsearch.Addresses)
	assert.Equal(t, "elastic", cfg.Elasticsearch.Username)
	assert.Equal(t, "testpass", cfg.Elasticsearch.Password)
	assert.Equal(t, "videos_test", cfg.Elasticsearch.Index)
	assert.Equal(t, 5, cfg.Elasticsearch.Shards)
	assert.Equal(t, 2, cfg.Elasticsearch.Replicas)
	assert.Contains(t, cfg.Kafka.ConsumeTopics, "video.metadata.created")
	assert.Equal(t, "search-test-group", cfg.Kafka.GroupID)
}

func TestLoader_DefaultValues(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	minimalConfig := `
elasticsearch:
  addresses:
    - http://localhost:9200

kafka:
  brokers:
    - localhost:9092
`
	err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := Loader()

	assert.Equal(t, 50056, cfg.Server.Port)
	assert.Equal(t, "videos", cfg.Elasticsearch.Index)
	assert.Equal(t, 3, cfg.Elasticsearch.Shards)
	assert.Equal(t, 1, cfg.Elasticsearch.Replicas)
	assert.Contains(t, cfg.Kafka.ConsumeTopics, "video.metadata.created")
	assert.Equal(t, "search-service-group", cfg.Kafka.GroupID)
}

// Note: Tests for missing required fields, invalid YAML, and non-existent config files
// are skipped because Loader() calls log.Fatalf which does os.Exit, not panic.
// These scenarios would terminate the test process, making them unsuitable for unit testing.
// In production, these errors are correctly handled by terminating the service.
