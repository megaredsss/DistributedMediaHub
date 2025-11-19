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

// Config holds the complete configuration for search service
type Config struct {
	Server        `mapstructure:"server"`
	Elasticsearch `mapstructure:"elasticsearch"`
	Kafka         `mapstructure:"kafka"`
	Logging       `mapstructure:"logging"`
}

// Server contains HTTP server configuration
type Server struct {
	Port          int           `mapstructure:"port"`
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderByte int           `mapstructure:"max_header_bytes"`
}

// Elasticsearch contains Elasticsearch configuration for full-text search
type Elasticsearch struct {
	Addresses []string `mapstructure:"addresses" validate:"required"` // ES cluster addresses
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
	Index     string   `mapstructure:"index"` // Index name for videos
	Shards    int      `mapstructure:"shards"`
	Replicas  int      `mapstructure:"replicas"`
}

// Kafka contains Kafka configuration for event consuming
type Kafka struct {
	Brokers       []string `mapstructure:"brokers" validate:"required"`
	ConsumeTopics []string `mapstructure:"consume_topics"` // video.metadata.created, video.metadata.updated, video.deleted
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
		log.Println("[search-service] Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("[search-service] Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("[search-service] Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	// Set default values for server configuration
	viper.SetDefault("server.port", 50056)
	viper.SetDefault("server.read_timeout", "20s")
	viper.SetDefault("server.write_timeout", "20s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Set default values for Elasticsearch configuration
	viper.SetDefault("elasticsearch.addresses", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.index", "videos")
	viper.SetDefault("elasticsearch.shards", 3)
	viper.SetDefault("elasticsearch.replicas", 1)

	// Set default values for Kafka configuration
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.consume_topics", []string{"video.metadata.created", "video.metadata.updated", "video.deleted"})
	viper.SetDefault("kafka.group_id", "search-service-group")

	// Set default values for logging configuration
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[search-service] Error reading config file: %v\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[search-service] Error unmarshal config: %v\n", err)
	}
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("[search-service] Missing required attributes: %v\n", err)
	}
	return &cfg
}
