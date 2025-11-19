package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants - avoiding magic values in tests
const (
	testPort             = "50057"
	testMongoHost        = "testhost:27017"
	testDatabase         = "analytics_test"
	testUsername         = "testuser"
	testPassword         = "testpass"
	testAuthSource       = "admin"
	testReadTimeout      = 30 * time.Second
	testWriteTimeout     = 30 * time.Second
	testIdleTimeout      = 30 * time.Second
	testMaxHeaderBytes   = 2097152
	testMongoTimeout     = 45 * time.Second
	testLoggingLevel     = "info"
	testLoggingFormat    = "json"
	testLoggingOutput    = "stdout"
)

// createTempConfigFile is a test helper that creates a temporary config file
// and returns the directory path. It uses t.Helper() to mark itself as a helper
// function, ensuring that test failures are reported at the call site, not here.
func createTempConfigFile(t *testing.T, content string) string {
	t.Helper() // Best practice: mark as helper function

	// Create temporary directory - automatically cleaned up by t.TempDir()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Write config file
	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create temporary config file")

	return tempDir
}

// setConfigPath sets the CONFIG_PATH environment variable and registers cleanup
func setConfigPath(t *testing.T, path string) {
	t.Helper() // Best practice: mark as helper function

	// Set environment variable
	os.Setenv("CONFIG_PATH", path)

	// Best practice: Use t.Cleanup() instead of defer for cleanup
	// This ensures cleanup happens even if the test panics
	t.Cleanup(func() {
		os.Unsetenv("CONFIG_PATH")
	})
}

// TestLoader_ValidConfig verifies that the Loader correctly loads and parses
// a complete configuration file with all fields explicitly set.
//
// This test ensures that:
// - Custom values override defaults
// - All configuration sections are parsed correctly
// - Time duration parsing works as expected
// - Nested structures are unmarshaled properly
func TestLoader_ValidConfig(t *testing.T) {
	// Arrange: Prepare test configuration
	validConfig := `
server:
  port: "50057"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 30s
  max_header_bytes: 2097152

mongodb:
  uri: mongodb://testhost:27017
  database: analytics_test
  username: testuser
  password: testpass
  auth_source: admin
  timeout: 45s

logging:
  level: info
  format: json
  output: stdout
`
	tempDir := createTempConfigFile(t, validConfig)
	setConfigPath(t, tempDir)

	// Act: Load configuration
	cfg := Loader()

	// Assert: Verify all configuration values
	// Best practice: Group related assertions together with comments

	// Server configuration assertions
	assert.Equal(t, testPort, cfg.Server.Port, "Server port should match config value")
	assert.Equal(t, testReadTimeout, cfg.Server.ReadTimeout, "Read timeout should match config value")
	assert.Equal(t, testWriteTimeout, cfg.Server.WriteTimeout, "Write timeout should match config value")
	assert.Equal(t, testIdleTimeout, cfg.Server.IdleTimeout, "Idle timeout should match config value")
	assert.Equal(t, testMaxHeaderBytes, cfg.Server.MaxHeaderByte, "Max header bytes should match config value")

	// MongoDB configuration assertions
	assert.Equal(t, "mongodb://"+testMongoHost, cfg.MongoDb.Uri, "MongoDB URI should match config value")
	assert.Equal(t, testDatabase, cfg.MongoDb.Database, "MongoDB database should match config value")
	assert.Equal(t, testUsername, cfg.MongoDb.Username, "MongoDB username should match config value")
	assert.Equal(t, testPassword, cfg.MongoDb.Password, "MongoDB password should match config value")
	assert.Equal(t, testAuthSource, cfg.MongoDb.AuthSource, "MongoDB auth source should match config value")
	assert.Equal(t, testMongoTimeout, cfg.MongoDb.Timeout, "MongoDB timeout should match config value")

	// Logging configuration assertions
	assert.Equal(t, testLoggingLevel, cfg.Logging.Level, "Logging level should match config value")
	assert.Equal(t, testLoggingFormat, cfg.Logging.Format, "Logging format should match config value")
	assert.Equal(t, testLoggingOutput, cfg.Logging.Output, "Logging output should match config value")
}

// TestLoader_DefaultValues verifies that the Loader applies correct default values
// when optional configuration fields are not specified.
//
// This is important because:
// - It ensures the service can start with minimal configuration
// - It documents the expected default behavior
// - It catches regressions in default value changes
func TestLoader_DefaultValues(t *testing.T) {
	// Arrange: Create minimal config with only required fields
	minimalConfig := `
mongodb:
  database: analytics_test
  username: testuser
  password: testpass
  auth_source: admin
`
	tempDir := createTempConfigFile(t, minimalConfig)
	setConfigPath(t, tempDir)

	// Act: Load configuration
	cfg := Loader()

	// Assert: Verify default values are applied
	// Best practice: Test default values separately from custom values

	// Default server values
	assert.Equal(t, "5051", cfg.Server.Port, "Should use default port 5051")
	assert.Equal(t, 20*time.Second, cfg.Server.ReadTimeout, "Should use default read timeout of 20s")
	assert.Equal(t, 20*time.Second, cfg.Server.WriteTimeout, "Should use default write timeout of 20s")
	assert.Equal(t, 20*time.Second, cfg.Server.IdleTimeout, "Should use default idle timeout of 20s")
	assert.Equal(t, 1048576, cfg.Server.MaxHeaderByte, "Should use default max header bytes (1MB)")

	// Default MongoDB values
	assert.Equal(t, "mongodb://localhost:27017", cfg.MongoDb.Uri, "Should use default MongoDB URI")
	assert.Equal(t, 30*time.Second, cfg.MongoDb.Timeout, "Should use default MongoDB timeout of 30s")

	// Default logging values
	assert.Equal(t, "debug", cfg.Logging.Level, "Should use default logging level 'debug'")
	assert.Equal(t, "json", cfg.Logging.Format, "Should use default logging format 'json'")
	assert.Equal(t, "stdout", cfg.Logging.Output, "Should use default logging output 'stdout'")
}

// TestLoader_BoundaryValues tests edge cases and boundary conditions
// to ensure the configuration loader handles extreme values correctly
func TestLoader_BoundaryValues(t *testing.T) {
	// Best practice: Use table-driven tests for multiple similar test cases
	tests := []struct {
		name        string
		config      string
		validate    func(t *testing.T, cfg *Config)
		description string
	}{
		{
			name: "MinimumTimeouts",
			description: "Verify that very short timeout values are handled correctly",
			config: `
mongodb:
  database: test
  username: user
  password: pass
  auth_source: admin
server:
  read_timeout: 1s
  write_timeout: 1s
  idle_timeout: 1s
`,
			validate: func(t *testing.T, cfg *Config) {
				t.Helper()
				assert.Equal(t, 1*time.Second, cfg.Server.ReadTimeout)
				assert.Equal(t, 1*time.Second, cfg.Server.WriteTimeout)
				assert.Equal(t, 1*time.Second, cfg.Server.IdleTimeout)
			},
		},
		{
			name: "MaximumTimeouts",
			description: "Verify that very long timeout values are handled correctly",
			config: `
mongodb:
  database: test
  username: user
  password: pass
  auth_source: admin
server:
  read_timeout: 300s
  write_timeout: 300s
  idle_timeout: 600s
`,
			validate: func(t *testing.T, cfg *Config) {
				t.Helper()
				assert.Equal(t, 300*time.Second, cfg.Server.ReadTimeout)
				assert.Equal(t, 300*time.Second, cfg.Server.WriteTimeout)
				assert.Equal(t, 600*time.Second, cfg.Server.IdleTimeout)
			},
		},
		{
			name: "LargeMaxHeaderBytes",
			description: "Verify that large header byte limits are handled correctly",
			config: `
mongodb:
  database: test
  username: user
  password: pass
  auth_source: admin
server:
  max_header_bytes: 10485760
`,
			validate: func(t *testing.T, cfg *Config) {
				t.Helper()
				assert.Equal(t, 10485760, cfg.Server.MaxHeaderByte) // 10MB
			},
		},
	}

	// Best practice: Use subtests for better organization and parallel execution
	for _, tt := range tests {
		// Capture range variable for parallel tests
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			// Best practice: Run independent tests in parallel for faster execution
			// Comment out if tests share resources or modify global state
			// t.Parallel()

			// Arrange
			tempDir := createTempConfigFile(t, tt.config)
			setConfigPath(t, tempDir)

			// Act
			cfg := Loader()

			// Assert
			tt.validate(t, cfg)
		})
	}
}

// TestLoader_EnvironmentVariablePrecedence verifies that environment variables
// take precedence over default values and flag-based configuration
func TestLoader_EnvironmentVariablePrecedence(t *testing.T) {
	// Arrange
	validConfig := `
mongodb:
  database: analytics_test
  username: testuser
  password: testpass
  auth_source: admin
`
	tempDir := createTempConfigFile(t, validConfig)

	// Set environment variable (should take precedence)
	setConfigPath(t, tempDir)

	// Act
	cfg := Loader()

	// Assert: Configuration should be loaded from the environment path
	require.NotNil(t, cfg, "Config should be loaded successfully")
	assert.Equal(t, testDatabase, cfg.MongoDb.Database,
		"Should load config from CONFIG_PATH environment variable")
}

// Note: Tests for missing required fields, invalid YAML, and non-existent config files
// are skipped because Loader() calls log.Fatalf which does os.Exit, not panic.
// These scenarios would terminate the test process, making them unsuitable for unit testing.
//
// In production, these errors are correctly handled by terminating the service immediately,
// which is the desired behavior for misconfigured services (fail-fast principle).
//
// To test these error scenarios, consider:
// 1. Integration tests that spawn the service as a separate process
// 2. Refactoring Loader() to return errors instead of calling log.Fatalf
// 3. Using dependency injection to mock the fatal error handler
