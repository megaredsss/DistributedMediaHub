package config

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config holds the complete configuration for streaming service
type Config struct {
	Server    `mapstructure:"server"`
	MinIO     `mapstructure:"minio"`
	Redis     `mapstructure:"redis"`
	Streaming `mapstructure:"streaming"`
	Kafka     `mapstructure:"kafka"`
	Logging   `mapstructure:"logging"`
}

// Server contains HTTP server configuration
type Server struct {
	Port          int           `mapstructure:"port"`
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderByte int           `mapstructure:"max_header_bytes"`
}

// MinIO contains MinIO/S3 storage configuration for video segments
type MinIO struct {
	Endpoint         string `mapstructure:"endpoint" validate:"required"`
	AccessKey        string `mapstructure:"access_key" validate:"required"`
	SecretKey        string `mapstructure:"secret_key" validate:"required"`
	UseSSL           bool   `mapstructure:"use_ssl"`
	BucketHLS        string `mapstructure:"bucket_hls"`        // hls-segments bucket
	BucketThumbnails string `mapstructure:"bucket_thumbnails"` // thumbnails bucket
}

// Redis contains Redis configuration for playlist caching
type Redis struct {
	Host        string        `mapstructure:"host" validate:"required"`
	Port        int           `mapstructure:"port" validate:"required"`
	Password    string        `mapstructure:"password"`
	DB          int           `mapstructure:"db"`
	PlaylistTTL time.Duration `mapstructure:"playlist_ttl"` // Cache TTL for playlists
	SegmentTTL  time.Duration `mapstructure:"segment_ttl"`  // Cache TTL for segment URLs
}

// Streaming contains HLS/DASH streaming configuration
type Streaming struct {
	EnableHLS  bool `mapstructure:"enable_hls"`  // Enable HLS streaming
	EnableDASH bool `mapstructure:"enable_dash"` // Enable DASH streaming (future)
	CDNEnabled bool `mapstructure:"cdn_enabled"` // Enable CDN integration
	CDNBaseURL string `mapstructure:"cdn_base_url"` // CDN base URL if enabled
}

// Kafka contains Kafka configuration for event publishing/consuming
type Kafka struct {
	Brokers       []string `mapstructure:"brokers" validate:"required"`
	PublishTopic  string   `mapstructure:"publish_topic"`  // video.viewed
	ConsumeTopics []string `mapstructure:"consume_topics"` // video.transcoded
	GroupID       string   `mapstructure:"group_id"`
}

// Logging contains logging configuration
type Logging struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, text
	Output string `mapstructure:"output"` // stdout, stderr, file path
}

// Loader loads configuration from file and environment variables
func Loader() *Config {
	envConfig := os.Getenv("CONFIG_PATH")

	// Check if flag already exists (for testing)
	var flagConfig *string
	if flag.Lookup("configPath") == nil {
		flagConfig = flag.String("configPath", "configs/dev", "config path")
	} else {
		val := flag.Lookup("configPath").Value.String()
		flagConfig = &val
	}

	if !flag.Parsed() {
		flag.Parse()
	}

	var configPath string

	fmt.Println(*flagConfig)
	if envConfig == "" {
		log.Println("[streaming-service] Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("[streaming-service] Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("[streaming-service] Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	// Set default values for server configuration
	viper.SetDefault("server.port", 50055)
	viper.SetDefault("server.read_timeout", "20s")
	viper.SetDefault("server.write_timeout", "20s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Set default values for MinIO configuration
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("minio.bucket_hls", "hls-segments")
	viper.SetDefault("minio.bucket_thumbnails", "thumbnails")

	// Set default values for Redis configuration
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.playlist_ttl", "1h")
	viper.SetDefault("redis.segment_ttl", "24h")

	// Set default values for streaming configuration
	viper.SetDefault("streaming.enable_hls", true)
	viper.SetDefault("streaming.enable_dash", false)
	viper.SetDefault("streaming.cdn_enabled", false)
	viper.SetDefault("streaming.cdn_base_url", "")

	// Set default values for Kafka configuration
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.publish_topic", "video.viewed")
	viper.SetDefault("kafka.consume_topics", []string{"video.transcoded"})
	viper.SetDefault("kafka.group_id", "streaming-service-group")

	// Set default values for logging configuration
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[streaming-service] Error reading config file: %v\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[streaming-service] Error unmarshal config: %v\n", err)
	}
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("[streaming-service] Missing required attributes: %v\n", err)
	}
	return &cfg
}
