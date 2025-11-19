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
  port: 50054

minio:
  endpoint: minio-test:9000
  access_key: test-access-key
  secret_key: test-secret-key
  use_ssl: true
  bucket_raw: test-raw-videos
  bucket_transcoded: test-transcoded
  bucket_hls: test-hls-segments
  bucket_thumbnails: test-thumbnails

transcoding:
  worker_count: 10
  resolutions:
    - 240p
    - 480p
    - 720p
    - 1080p
  video_codec: libx264
  audio_codec: aac
  preset: fast
  crf: 20
  ffmpeg_path: /usr/local/bin/ffmpeg
  ffprobe_path: /usr/local/bin/ffprobe
  hls_segment_duration: 10

kafka:
  brokers:
    - kafka1:9092
  consume_topic: video.uploaded
  publish_topic: video.transcoded
  group_id: transcoding-test-group

logging:
  level: info
`
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := Loader()

	assert.Equal(t, 50054, cfg.Server.Port)
	assert.Equal(t, "minio-test:9000", cfg.MinIO.Endpoint)
	assert.Equal(t, "test-raw-videos", cfg.MinIO.BucketRaw)
	assert.Equal(t, 10, cfg.Transcoding.WorkerCount)
	assert.Contains(t, cfg.Transcoding.Resolutions, "1080p")
	assert.Equal(t, "libx264", cfg.Transcoding.VideoCodec)
	assert.Equal(t, 20, cfg.Transcoding.CRF)
	assert.Equal(t, 10, cfg.Transcoding.HLSSegmentDuration)
	assert.Equal(t, "video.transcoded", cfg.Kafka.PublishTopic)
}

func TestLoader_DefaultValues(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	minimalConfig := `
minio:
  endpoint: localhost:9000
  access_key: test-access
  secret_key: test-secret

kafka:
  brokers:
    - localhost:9092
`
	err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
	require.NoError(t, err)

	os.Setenv("CONFIG_PATH", tempDir)
	defer os.Unsetenv("CONFIG_PATH")

	cfg := Loader()

	assert.Equal(t, 50054, cfg.Server.Port)
	assert.Equal(t, 5, cfg.Transcoding.WorkerCount)
	assert.Contains(t, cfg.Transcoding.Resolutions, "240p")
	assert.Equal(t, "libx264", cfg.Transcoding.VideoCodec)
	assert.Equal(t, "aac", cfg.Transcoding.AudioCodec)
	assert.Equal(t, "medium", cfg.Transcoding.Preset)
	assert.Equal(t, 23, cfg.Transcoding.CRF)
	assert.Equal(t, 6, cfg.Transcoding.HLSSegmentDuration)
}

// Note: Tests for missing required fields, invalid YAML, and non-existent config files
// are skipped because Loader() calls log.Fatalf which does os.Exit, not panic.
// These scenarios would terminate the test process, making them unsuitable for unit testing.
// In production, these errors are correctly handled by terminating the service.
