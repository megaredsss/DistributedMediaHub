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

// Config holds the complete configuration for analytics service
type Config struct {
	Server  `mapstructure:"server"`
	MongoDb `mapstructure:"mongodb"`
	Logging `mapstructure:"logging"`
}

// Server contains HTTP server configuration
type Server struct {
	Port          string
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderByte int           `mapstructure:"max_header_bytes"`
}

// MongoDb contains MongoDB connection configuration for time-series analytics data
type MongoDb struct {
	Uri        string        `mapstructure:"uri"`
	Database   string        `mapstructure:"database" validate:"required"`
	Username   string        `mapstructure:"username" validate:"required"`
	Password   string        `mapstructure:"password" validate:"required"`
	AuthSource string        `mapstructure:"auth_source" validate:"required"`
	Timeout    time.Duration `mapstructure:"timeout"`
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
		// Flag already registered, get its value
		val := flag.Lookup("configPath").Value.String()
		flagConfig = &val
	}

	if !flag.Parsed() {
		flag.Parse()
	}

	var configPath string

	fmt.Println(*flagConfig)
	if envConfig == "" {
		log.Println("[analytics-service] Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("[analytics-service] Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("[analytics-service] Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	// Set default values for server configuration
	viper.SetDefault("server.port", "5051")
	viper.SetDefault("server.read_timeout", "20s")
	viper.SetDefault("server.write_timeout", "20s")
	viper.SetDefault("server.idle_timeout", "20s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Set default values for MongoDB configuration
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.timeout", "30s")

	// Set default values for logging configuration
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[analytics-service] Error reading config file: %v\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[analytics-service] Error unmarshal config: %v\n", err)
	}
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("[analytics-service] Missing required attributes: %v\n", err)
	}
	return &cfg
}
