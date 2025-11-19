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
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 30s
  max_header_bytes: 2097152

postgresql:
  host: testhost
  port: 5433
  user: testuser
  password: testpass
  name: auth_test

jwt:
  secret_key: test-secret-key-min-32-characters-long
  access_token_expire: 30m
  refresh_token_expire: 336h
  issuer: test-mediahub
  algorithm: HS256

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
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, 2097152, cfg.Server.MaxHeaderByte)

	// Assert PostgreSQL configuration
	assert.Equal(t, "testhost", cfg.PostgreSQL.Host)
	assert.Equal(t, 5433, cfg.PostgreSQL.Port)
	assert.Equal(t, "testuser", cfg.PostgreSQL.User)
	assert.Equal(t, "testpass", cfg.PostgreSQL.Password)
	assert.Equal(t, "auth_test", cfg.PostgreSQL.Name)

	// Assert JWT configuration
	assert.Equal(t, "test-secret-key-min-32-characters-long", cfg.JWT.SecretKey)
	assert.Equal(t, 30*time.Minute, cfg.JWT.AccessTokenExpire)
	assert.Equal(t, 336*time.Hour, cfg.JWT.RefreshTokenExpire)
	assert.Equal(t, "test-mediahub", cfg.JWT.Issuer)
	assert.Equal(t, "HS256", cfg.JWT.Algorithm)

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
postgresql:
  user: testuser
  password: testpass
  name: auth_test

jwt:
  secret_key: test-secret-key-min-32-characters-long
`
	err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
	require.NoError(t, err)

	// Set environment variable
	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	// Load config
	cfg := Loader()

	// Assert default server values
	assert.Equal(t, 20*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 20*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 20*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, 1048576, cfg.Server.MaxHeaderByte)

	// Assert default PostgreSQL values
	assert.Equal(t, "localhost", cfg.PostgreSQL.Host)
	assert.Equal(t, 5432, cfg.PostgreSQL.Port)

	// Assert default JWT values
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTokenExpire)
	assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshTokenExpire)
	assert.Equal(t, "mediahub", cfg.JWT.Issuer)
	assert.Equal(t, "HS256", cfg.JWT.Algorithm)

	// Assert default logging values
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
}

// Note: Tests for missing required fields, invalid YAML, and non-existent config files
// are skipped because Loader() calls log.Fatalf which does os.Exit, not panic.
// These scenarios would terminate the test process, making them unsuitable for unit testing.
// In production, these errors are correctly handled by terminating the service.
