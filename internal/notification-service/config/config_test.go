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
  port: 50058

smtp:
  host: smtp.test.com
  port: 465
  username: test@example.com
  password: testpass
  from: noreply@example.com
  from_name: TestMediaHub

notification:
  enable_email: true
  enable_websocket: true
  template_dir: /templates/test
  max_retries: 5
  retry_delay: 10s

kafka:
  brokers:
    - kafka1:9092
  consume_topics:
    - user.registered
    - video.transcoded
    - notification.send
  group_id: notification-test-group

logging:
  level: info
`
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := Loader()

	assert.Equal(t, 50058, cfg.Server.Port)
	assert.Equal(t, "smtp.test.com", cfg.SMTP.Host)
	assert.Equal(t, 465, cfg.SMTP.Port)
	assert.Equal(t, "test@example.com", cfg.SMTP.Username)
	assert.Equal(t, "testpass", cfg.SMTP.Password)
	assert.Equal(t, "noreply@example.com", cfg.SMTP.From)
	assert.Equal(t, "TestMediaHub", cfg.SMTP.FromName)
	assert.True(t, cfg.Notification.EnableEmail)
	assert.True(t, cfg.Notification.EnableWebSocket)
	assert.Equal(t, "/templates/test", cfg.Notification.TemplateDir)
	assert.Equal(t, 5, cfg.Notification.MaxRetries)
	assert.Equal(t, 10*time.Second, cfg.Notification.RetryDelay)
	assert.Contains(t, cfg.Kafka.ConsumeTopics, "user.registered")
}

func TestLoader_DefaultValues(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	minimalConfig := `
smtp:
  host: smtp.gmail.com
  port: 587
  username: test@example.com
  password: testpass
  from: noreply@example.com

kafka:
  brokers:
    - localhost:9092
`
	err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := Loader()

	assert.Equal(t, 50058, cfg.Server.Port)
	assert.Equal(t, "MediaHub", cfg.SMTP.FromName)
	assert.True(t, cfg.Notification.EnableEmail)
	assert.True(t, cfg.Notification.EnableWebSocket)
	assert.Equal(t, "templates/email", cfg.Notification.TemplateDir)
	assert.Equal(t, 3, cfg.Notification.MaxRetries)
	assert.Equal(t, 5*time.Second, cfg.Notification.RetryDelay)
	assert.Contains(t, cfg.Kafka.ConsumeTopics, "notification.send")
	assert.Equal(t, "notification-service-group", cfg.Kafka.GroupID)
}

// Note: Tests for missing required fields, invalid YAML, and non-existent config files
// are skipped because Loader() calls log.Fatalf which does os.Exit, not panic.
// These scenarios would terminate the test process, making them unsuitable for unit testing.
// In production, these errors are correctly handled by terminating the service.
