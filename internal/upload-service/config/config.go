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

// Config holds the complete configuration for upload service
type Config struct {
	Server  `mapstructure:"server"`
	MinIO   `mapstructure:"minio"`
	Redis   `mapstructure:"redis"`
	Upload  `mapstructure:"upload"`
	Kafka   `mapstructure:"kafka"`
	Logging `mapstructure:"logging"`
}

// Server contains HTTP server configuration
type Server struct {
	Port          int           `mapstructure:"port"`
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderByte int           `mapstructure:"max_header_bytes"`
}

// MinIO contains MinIO/S3 storage configuration
type MinIO struct {
	Endpoint        string `mapstructure:"endpoint" validate:"required"`
	AccessKey       string `mapstructure:"access_key" validate:"required"`
	SecretKey       string `mapstructure:"secret_key" validate:"required"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketRaw       string `mapstructure:"bucket_raw"`       // raw-videos bucket
	BucketTemp      string `mapstructure:"bucket_temp"`      // temp-uploads bucket
}

// Redis contains Redis configuration for upload progress tracking
type Redis struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// Upload contains upload-specific configuration
type Upload struct {
	MaxFileSize        int64    `mapstructure:"max_file_size"`        // in bytes, default 5GB
	ChunkSize          int      `mapstructure:"chunk_size"`           // in bytes, default 5MB
	AllowedExtensions  []string `mapstructure:"allowed_extensions"`
	TempDir            string   `mapstructure:"temp_dir"`
	CleanupInterval    time.Duration `mapstructure:"cleanup_interval"` // cleanup failed uploads
}

// Kafka contains Kafka configuration for event publishing
type Kafka struct {
	Brokers []string `mapstructure:"brokers" validate:"required"`
	Topic   string   `mapstructure:"topic"`
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
		log.Println("[upload-service] Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("[upload-service] Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("[upload-service] Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	// Set default values for server configuration
	viper.SetDefault("server.port", 50053)
	viper.SetDefault("server.read_timeout", "20s")
	viper.SetDefault("server.write_timeout", "20s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Set default values for MinIO configuration
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("minio.bucket_raw", "raw-videos")
	viper.SetDefault("minio.bucket_temp", "temp-uploads")

	// Set default values for Redis configuration
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)

	// Set default values for upload configuration
	viper.SetDefault("upload.max_file_size", 5368709120) // 5GB
	viper.SetDefault("upload.chunk_size", 5242880)       // 5MB
	viper.SetDefault("upload.allowed_extensions", []string{"mp4", "avi", "mov", "mkv", "webm", "flv"})
	viper.SetDefault("upload.temp_dir", "/tmp/uploads")
	viper.SetDefault("upload.cleanup_interval", "1h")

	// Set default values for Kafka configuration
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.topic", "video.uploaded")

	// Set default values for logging configuration
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[upload-service] Error reading config file: %v\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[upload-service] Error unmarshal config: %v\n", err)
	}
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("[upload-service] Missing required attributes: %v\n", err)
	}
	return &cfg
}
