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

// Config holds the complete configuration for notification service
type Config struct {
	Server       `mapstructure:"server"`
	SMTP         `mapstructure:"smtp"`
	Notification `mapstructure:"notification"`
	Kafka        `mapstructure:"kafka"`
	Logging      `mapstructure:"logging"`
}

// Server contains gRPC server configuration
type Server struct {
	Port          int           `mapstructure:"port"`
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderByte int           `mapstructure:"max_header_bytes"`
}

// SMTP contains email server configuration
type SMTP struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required"`
	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	From     string `mapstructure:"from" validate:"required"` // Sender email address
	FromName string `mapstructure:"from_name"`                // Sender name
}

// Notification contains notification settings
type Notification struct {
	EnableEmail     bool   `mapstructure:"enable_email"`      // Enable email notifications
	EnableWebSocket bool   `mapstructure:"enable_websocket"`  // Enable WebSocket notifications
	TemplateDir     string `mapstructure:"template_dir"`      // Path to email templates
	MaxRetries      int    `mapstructure:"max_retries"`       // Max retry attempts for failed notifications
	RetryDelay      time.Duration `mapstructure:"retry_delay"` // Delay between retries
}

// Kafka contains Kafka configuration for event consuming
type Kafka struct {
	Brokers       []string `mapstructure:"brokers" validate:"required"`
	ConsumeTopics []string `mapstructure:"consume_topics"` // user.registered, video.transcoded, notification.send
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
		log.Println("[notification-service] Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("[notification-service] Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("[notification-service] Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	// Set default values for server configuration
	viper.SetDefault("server.port", 50058)
	viper.SetDefault("server.read_timeout", "20s")
	viper.SetDefault("server.write_timeout", "20s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Set default values for SMTP configuration
	viper.SetDefault("smtp.host", "smtp.gmail.com")
	viper.SetDefault("smtp.port", 587)
	viper.SetDefault("smtp.from_name", "MediaHub")

	// Set default values for notification configuration
	viper.SetDefault("notification.enable_email", true)
	viper.SetDefault("notification.enable_websocket", true)
	viper.SetDefault("notification.template_dir", "templates/email")
	viper.SetDefault("notification.max_retries", 3)
	viper.SetDefault("notification.retry_delay", "5s")

	// Set default values for Kafka configuration
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.consume_topics", []string{"user.registered", "video.transcoded", "notification.send"})
	viper.SetDefault("kafka.group_id", "notification-service-group")

	// Set default values for logging configuration
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[notification-service] Error reading config file: %v\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[notification-service] Error unmarshal config: %v\n", err)
	}
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("[notification-service] Missing required attributes: %v\n", err)
	}
	return &cfg
}
