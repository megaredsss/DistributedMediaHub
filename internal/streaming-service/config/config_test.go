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
  port: 50055

minio:
  endpoint: minio-test:9000
  access_key: test-access-key
  secret_key: test-secret-key
  use_ssl: true
  bucket_hls: test-hls-segments
  bucket_thumbnails: test-thumbnails

redis:
  host: redis-test
  port: 6380
  password: testpass
  db: 1
  playlist_ttl: 2h
  segment_ttl: 48h

streaming:
  enable_hls: true
  enable_dash: true
  cdn_enabled: true
  cdn_base_url: https://cdn.example.com

kafka:
  brokers:
    - kafka1:9092
  publish_topic: video.viewed
  consume_topics:
    - video.transcoded
  group_id: streaming-test-group

logging:
  level: info
`
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := Loader()

	assert.Equal(t, 50055, cfg.Server.Port)
	assert.Equal(t, "minio-test:9000", cfg.MinIO.Endpoint)
	assert.Equal(t, "test-hls-segments", cfg.MinIO.BucketHLS)
	assert.Equal(t, 2*time.Hour, cfg.Redis.PlaylistTTL)
	assert.Equal(t, 48*time.Hour, cfg.Redis.SegmentTTL)
	assert.True(t, cfg.Streaming.EnableHLS)
	assert.True(t, cfg.Streaming.EnableDASH)
	assert.True(t, cfg.Streaming.CDNEnabled)
	assert.Equal(t, "https://cdn.example.com", cfg.Streaming.CDNBaseURL)
	assert.Equal(t, "video.viewed", cfg.Kafka.PublishTopic)
}

func TestLoader_DefaultValues(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	minimalConfig := `
minio:
  endpoint: localhost:9000
  access_key: test-access
  secret_key: test-secret

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

	assert.Equal(t, 50055, cfg.Server.Port)
	assert.Equal(t, "hls-segments", cfg.MinIO.BucketHLS)
	assert.Equal(t, "thumbnails", cfg.MinIO.BucketThumbnails)
	assert.Equal(t, 1*time.Hour, cfg.Redis.PlaylistTTL)
	assert.Equal(t, 24*time.Hour, cfg.Redis.SegmentTTL)
	assert.True(t, cfg.Streaming.EnableHLS)
	assert.False(t, cfg.Streaming.EnableDASH)
	assert.False(t, cfg.Streaming.CDNEnabled)
}

// Note: Tests for missing required fields, invalid YAML, and non-existent config files
// are skipped because Loader() calls log.Fatalf which does os.Exit, not panic.
// These scenarios would terminate the test process, making them unsuitable for unit testing.
// In production, these errors are correctly handled by terminating the service.
