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
  port: "9090"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 90s
  max_header_bytes: 2097152

redis:
  host: redis-test
  port: 6380
  password: testpass
  db: 1

rate_limit:
  enabled: true
  requests_per_minute: 200
  burst_size: 40

cors:
  allowed_origins:
    - http://test1.com
    - http://test2.com
  allowed_methods:
    - GET
    - POST
  allowed_headers:
    - Content-Type

services:
  auth_service: localhost:50051
  video_service: localhost:50052
  upload_service: localhost:50053
  transcoding_service: localhost:50054
  streaming_service: localhost:50055
  search_service: localhost:50056
  analytics_service: localhost:50057
  notification_service: localhost:50058

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
	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 90*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, 2097152, cfg.Server.MaxHeaderByte)

	// Assert Redis configuration
	assert.Equal(t, "redis-test", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)
	assert.Equal(t, "testpass", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)

	// Assert rate limit configuration
	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 200, cfg.RateLimit.RequestsPerMinute)
	assert.Equal(t, 40, cfg.RateLimit.BurstSize)

	// Assert CORS configuration
	assert.Equal(t, []string{"http://test1.com", "http://test2.com"}, cfg.CORS.AllowedOrigins)
	assert.Equal(t, []string{"GET", "POST"}, cfg.CORS.AllowedMethods)
	assert.Equal(t, []string{"Content-Type"}, cfg.CORS.AllowedHeaders)

	// Assert services configuration
	assert.Equal(t, "localhost:50051", cfg.Services.AuthService)
	assert.Equal(t, "localhost:50052", cfg.Services.VideoService)
	assert.Equal(t, "localhost:50053", cfg.Services.UploadService)

	// Assert logging configuration
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
}

func TestLoader_DefaultValues(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create minimal config file with only required fields
	minimalConfig := `
redis:
  host: localhost
  port: 6379

services:
  auth_service: localhost:50051
  video_service: localhost:50052
  upload_service: localhost:50053
  transcoding_service: localhost:50054
  streaming_service: localhost:50055
  search_service: localhost:50056
  analytics_service: localhost:50057
  notification_service: localhost:50058
`
	err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
	require.NoError(t, err)

	// Set environment variable
	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	// Load config
	cfg := Loader()

	// Assert default server values
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 15*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, 1048576, cfg.Server.MaxHeaderByte)

	// Assert default Redis values
	assert.Equal(t, 0, cfg.Redis.DB)

	// Assert default rate limit values
	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 100, cfg.RateLimit.RequestsPerMinute)
	assert.Equal(t, 20, cfg.RateLimit.BurstSize)

	// Assert default CORS values
	assert.Contains(t, cfg.CORS.AllowedOrigins, "http://localhost:3000")
	assert.Contains(t, cfg.CORS.AllowedMethods, "GET")
	assert.Contains(t, cfg.CORS.AllowedHeaders, "Content-Type")

	// Assert default logging values
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
}

// Note: Tests for missing required fields, invalid YAML, and non-existent config files
// are skipped because Loader() calls log.Fatalf which does os.Exit, not panic.
// These scenarios would terminate the test process, making them unsuitable for unit testing.
// In production, these errors are correctly handled by terminating the service.
