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

// Config holds the complete configuration for API Gateway
type Config struct {
	Server    `mapstructure:"server"`
	Redis     `mapstructure:"redis"`
	RateLimit `mapstructure:"rate_limit"`
	CORS      `mapstructure:"cors"`
	Services  `mapstructure:"services"`
	Logging   `mapstructure:"logging"`
}

// Server contains HTTP server configuration
type Server struct {
	Port          string
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderByte int           `mapstructure:"max_header_bytes"`
}

// Redis contains Redis configuration for rate limiting and caching
type Redis struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// RateLimit contains rate limiting configuration
type RateLimit struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	BurstSize         int  `mapstructure:"burst_size"`
}

// CORS contains CORS configuration
type CORS struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// Services contains gRPC service endpoints
type Services struct {
	AuthService         string `mapstructure:"auth_service" validate:"required"`
	VideoService        string `mapstructure:"video_service" validate:"required"`
	UploadService       string `mapstructure:"upload_service" validate:"required"`
	TranscodingService  string `mapstructure:"transcoding_service" validate:"required"`
	StreamingService    string `mapstructure:"streaming_service" validate:"required"`
	SearchService       string `mapstructure:"search_service" validate:"required"`
	AnalyticsService    string `mapstructure:"analytics_service" validate:"required"`
	NotificationService string `mapstructure:"notification_service" validate:"required"`
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
		log.Println("[api-gateway] Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("[api-gateway] Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("[api-gateway] Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	// Set default values for server configuration
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Set default values for Redis configuration
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)

	// Set default values for rate limiting
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_minute", 100)
	viper.SetDefault("rate_limit.burst_size", 20)

	// Set default values for CORS
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:5173"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Content-Type", "Authorization"})

	// Set default values for services
	viper.SetDefault("services.auth_service", "localhost:50051")
	viper.SetDefault("services.video_service", "localhost:50052")
	viper.SetDefault("services.upload_service", "localhost:50053")
	viper.SetDefault("services.transcoding_service", "localhost:50054")
	viper.SetDefault("services.streaming_service", "localhost:50055")
	viper.SetDefault("services.search_service", "localhost:50056")
	viper.SetDefault("services.analytics_service", "localhost:50057")
	viper.SetDefault("services.notification_service", "localhost:50058")

	// Set default values for logging configuration
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[api-gateway] Error reading config file: %v\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[api-gateway] Error unmarshal config: %v\n", err)
	}
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("[api-gateway] Missing required attributes: %v\n", err)
	}
	return &cfg
}
