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

// Config holds the complete configuration for transcoding service
type Config struct {
	Server      `mapstructure:"server"`
	MinIO       `mapstructure:"minio"`
	Transcoding `mapstructure:"transcoding"`
	Kafka       `mapstructure:"kafka"`
	Logging     `mapstructure:"logging"`
}

// Server contains gRPC server configuration
type Server struct {
	Port          int           `mapstructure:"port"`
	ReadTimeout   time.Duration `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderByte int           `mapstructure:"max_header_bytes"`
}

// MinIO contains MinIO/S3 storage configuration
type MinIO struct {
	Endpoint          string `mapstructure:"endpoint" validate:"required"`
	AccessKey         string `mapstructure:"access_key" validate:"required"`
	SecretKey         string `mapstructure:"secret_key" validate:"required"`
	UseSSL            bool   `mapstructure:"use_ssl"`
	BucketRaw         string `mapstructure:"bucket_raw"`         // raw-videos bucket
	BucketTranscoded  string `mapstructure:"bucket_transcoded"`  // transcoded bucket
	BucketHLS         string `mapstructure:"bucket_hls"`         // hls-segments bucket
	BucketThumbnails  string `mapstructure:"bucket_thumbnails"`  // thumbnails bucket
}

// Transcoding contains FFmpeg and transcoding configuration
type Transcoding struct {
	WorkerCount   int      `mapstructure:"worker_count"`   // Number of parallel workers
	Resolutions   []string `mapstructure:"resolutions"`    // e.g., ["240p", "720p", "1080p"]
	VideoCodec    string   `mapstructure:"video_codec"`    // e.g., libx264
	AudioCodec    string   `mapstructure:"audio_codec"`    // e.g., aac
	Preset        string   `mapstructure:"preset"`         // e.g., medium, fast, slow
	CRF           int      `mapstructure:"crf"`            // Constant Rate Factor (23 recommended)
	FFmpegPath    string   `mapstructure:"ffmpeg_path"`    // Path to ffmpeg binary
	FFprobePath   string   `mapstructure:"ffprobe_path"`   // Path to ffprobe binary
	HLSSegmentDuration int `mapstructure:"hls_segment_duration"` // HLS segment duration in seconds
}

// Kafka contains Kafka configuration for event consuming/publishing
type Kafka struct {
	Brokers       []string `mapstructure:"brokers" validate:"required"`
	ConsumeTopic  string   `mapstructure:"consume_topic"`  // video.uploaded
	PublishTopic  string   `mapstructure:"publish_topic"`  // video.transcoded
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
		log.Println("[transcoding-service] Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("[transcoding-service] Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("[transcoding-service] Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	// Set default values for server configuration
	viper.SetDefault("server.port", 50054)
	viper.SetDefault("server.read_timeout", "20s")
	viper.SetDefault("server.write_timeout", "20s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Set default values for MinIO configuration
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("minio.bucket_raw", "raw-videos")
	viper.SetDefault("minio.bucket_transcoded", "transcoded")
	viper.SetDefault("minio.bucket_hls", "hls-segments")
	viper.SetDefault("minio.bucket_thumbnails", "thumbnails")

	// Set default values for transcoding configuration
	viper.SetDefault("transcoding.worker_count", 5)
	viper.SetDefault("transcoding.resolutions", []string{"240p", "720p", "1080p"})
	viper.SetDefault("transcoding.video_codec", "libx264")
	viper.SetDefault("transcoding.audio_codec", "aac")
	viper.SetDefault("transcoding.preset", "medium")
	viper.SetDefault("transcoding.crf", 23)
	viper.SetDefault("transcoding.ffmpeg_path", "/usr/bin/ffmpeg")
	viper.SetDefault("transcoding.ffprobe_path", "/usr/bin/ffprobe")
	viper.SetDefault("transcoding.hls_segment_duration", 6)

	// Set default values for Kafka configuration
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.consume_topic", "video.uploaded")
	viper.SetDefault("kafka.publish_topic", "video.transcoded")
	viper.SetDefault("kafka.group_id", "transcoding-service-group")

	// Set default values for logging configuration
	viper.SetDefault("logging.level", "debug")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("[transcoding-service] Error reading config file: %v\n", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[transcoding-service] Error unmarshal config: %v\n", err)
	}
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("[transcoding-service] Missing required attributes: %v\n", err)
	}
	return &cfg
}
