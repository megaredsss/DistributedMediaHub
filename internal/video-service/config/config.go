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

// Config holds the complete configuration for video metadata service
type Config struct {
	Server     `mapstructure:"server"`
	PostgreSQL `mapstructure:"postgresql"`
	Redis      `mapstructure:"redis"`
	Kafka      `mapstructure:"kafka"`
	Logging    `mapstructure:"logging"`
}

// Server contains gRPC server configuration
type Server struct {
	Port          int           `mapstructure:"port"`
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderByte int           `mapstructure:"max_header_bytes"`
}

// PostgreSQL contains PostgreSQL database configuration for video metadata
type PostgreSQL struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Name     string `mapstructure:"name" validate:"required"`
}

// Redis contains Redis configuration for caching video metadata
type Redis struct {
	Host     string        `mapstructure:"host" validate:"required"`
	Port     int           `mapstructure:"port" validate:"required"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	CacheTTL time.Duration `mapstructure:"cache_ttl"` // Cache TTL for video metadata
}

// Kafka contains Kafka configuration for event publishing/consuming
type Kafka struct {
	Brokers       []string `mapstructure:"brokers" validate:"required"`
	PublishTopic  string   `mapstructure:"publish_topic"`  // video.metadata.created
	ConsumeTopics []string `mapstructure:"consume_topics"` // video.uploaded, video.transcoded
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
		log.Println("[video-service] Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("[video-service] Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("[video-service] Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	// Set default values for server configuration
	viper.SetDefault("server.port", 50052)
	viper.SetDefault("server.read_timeout", "20s")
	viper.SetDefault("server.write_timeout", "20s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Set default values for PostgreSQL configuration
	viper.SetDefault("postgresql.host", "localhost")
	viper.SetDefault("postgresql.port", 5432)

	// Set default values for Redis configuration
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.cache_ttl", "1h")

	// Set default values for Kafka configuration
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.publish_topic", "video.metadata.created")
	viper.SetDefault("kafka.consume_topics", []string{"video.uploaded", "video.transcoded"})
	viper.SetDefault("kafka.group_id", "video-service-group")

	// Set default values for logging configuration
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[video-service] Error reading config file: %v\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[video-service] Error unmarshal config: %v\n", err)
	}
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("[video-service] Missing required attributes: %v\n", err)
	}
	return &cfg
}
