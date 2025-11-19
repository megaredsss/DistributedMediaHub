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

// Config holds the complete configuration for auth service
type Config struct {
	Server     `mapstructure:"server"`
	PostgreSQL `mapstructure:"postgresql"`
	JWT        `mapstructure:"jwt"`
	Logging    `mapstructure:"logging"`
}

// Server contains HTTP server configuration
type Server struct {
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderByte int           `mapstructure:"max_header_bytes"`
}

// PostgreSQL contains PostgreSQL database configuration for user data
type PostgreSQL struct {
	Host     string `mapstructure:"host" env-default:"localhost"`
	Port     int    `mapstructure:"port" env-default:"5432"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Name     string `mapstructure:"name" validate:"required"`
}

// JWT contains JWT token configuration for authentication
type JWT struct {
	SecretKey          string        `mapstructure:"secret_key" validate:"required"`
	AccessTokenExpire  time.Duration `mapstructure:"access_token_expire"`
	RefreshTokenExpire time.Duration `mapstructure:"refresh_token_expire"`
	Issuer             string        `mapstructure:"issuer"`
	Algorithm          string        `mapstructure:"algorithm"`
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
		log.Println("[auth-service] Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("[auth-service] Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("[auth-service] Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	// Set default values for server configuration
	viper.SetDefault("server.read_timeout", "20s")
	viper.SetDefault("server.write_timeout", "20s")
	viper.SetDefault("server.idle_timeout", "20s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Set default values for PostgreSQL configuration
	viper.SetDefault("postgresql.host", "localhost")
	viper.SetDefault("postgresql.port", 5432)

	// Set default values for JWT configuration
	viper.SetDefault("jwt.access_token_expire", "15m")
	viper.SetDefault("jwt.refresh_token_expire", "168h") // 7 days
	viper.SetDefault("jwt.issuer", "mediahub")
	viper.SetDefault("jwt.algorithm", "HS256")

	// Set default values for logging configuration
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[auth-service] Error reading config file: %v\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[auth-service] Error unmarshal config: %v\n", err)
	}
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("[auth-service] Missing required attributes: %v\n", err)
	}
	return &cfg
}
