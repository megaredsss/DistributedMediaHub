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
	// Create temporary config directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create valid config file
	validConfig := `
server:
  port: 8053
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 90s
  max_header_bytes: 2097152

minio:
  endpoint: minio-test:9000
  access_key: test-access-key
  secret_key: test-secret-key
  use_ssl: true
  bucket_raw: test-raw-videos
  bucket_temp: test-temp-uploads

redis:
  host: redis-test
  port: 6380
  password: testpass
  db: 1

upload:
  max_file_size: 10737418240
  chunk_size: 10485760
  allowed_extensions:
    - mp4
    - avi
  temp_dir: /tmp/test-uploads
  cleanup_interval: 2h

kafka:
  brokers:
    - kafka1:9092
    - kafka2:9092
  topic: test.video.uploaded

logging:
  level: info
  format: json
  output: stdout
`
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	// Set environment variable
	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	// Load config
	cfg := Loader()

	// Assert server configuration
	assert.Equal(t, 8053, cfg.Server.Port)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)

	// Assert MinIO configuration
	assert.Equal(t, "minio-test:9000", cfg.MinIO.Endpoint)
	assert.Equal(t, "test-access-key", cfg.MinIO.AccessKey)
	assert.Equal(t, "test-secret-key", cfg.MinIO.SecretKey)
	assert.True(t, cfg.MinIO.UseSSL)
	assert.Equal(t, "test-raw-videos", cfg.MinIO.BucketRaw)
	assert.Equal(t, "test-temp-uploads", cfg.MinIO.BucketTemp)

	// Assert Redis configuration
	assert.Equal(t, "redis-test", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)

	// Assert upload configuration
	assert.Equal(t, int64(10737418240), cfg.Upload.MaxFileSize)
	assert.Equal(t, 10485760, cfg.Upload.ChunkSize)
	assert.Equal(t, []string{"mp4", "avi"}, cfg.Upload.AllowedExtensions)
	assert.Equal(t, "/tmp/test-uploads", cfg.Upload.TempDir)
	assert.Equal(t, 2*time.Hour, cfg.Upload.CleanupInterval)

	// Assert Kafka configuration
	assert.Equal(t, []string{"kafka1:9092", "kafka2:9092"}, cfg.Kafka.Brokers)
	assert.Equal(t, "test.video.uploaded", cfg.Kafka.Topic)

	// Assert logging configuration
	assert.Equal(t, "info", cfg.Logging.Level)
}

func TestLoader_DefaultValues(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create minimal config file with only required fields
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

	// Set environment variable
	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	// Load config
	cfg := Loader()

	// Assert default server values
	assert.Equal(t, 50053, cfg.Server.Port)
	assert.Equal(t, 20*time.Second, cfg.Server.ReadTimeout)

	// Assert default MinIO values
	assert.False(t, cfg.MinIO.UseSSL)
	assert.Equal(t, "raw-videos", cfg.MinIO.BucketRaw)
	assert.Equal(t, "temp-uploads", cfg.MinIO.BucketTemp)

	// Assert default upload values
	assert.Equal(t, int64(5368709120), cfg.Upload.MaxFileSize) // 5GB
	assert.Equal(t, 5242880, cfg.Upload.ChunkSize)             // 5MB
	assert.Contains(t, cfg.Upload.AllowedExtensions, "mp4")
	assert.Equal(t, "/tmp/uploads", cfg.Upload.TempDir)
	assert.Equal(t, 1*time.Hour, cfg.Upload.CleanupInterval)

	// Assert default Kafka values
	assert.Equal(t, "video.uploaded", cfg.Kafka.Topic)

	// Assert default logging values
	assert.Equal(t, "debug", cfg.Logging.Level)
}

// Note: Tests for missing required fields, invalid YAML, and non-existent config files
// are skipped because Loader() calls log.Fatalf which does os.Exit, not panic.
// These scenarios would terminate the test process, making them unsuitable for unit testing.
// In production, these errors are correctly handled by terminating the service.
