package config

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Env         string `mapstructure:"env" env-default:"dev" env-required:"true"`
	HTTP_Server `mapstructure:"http_server"`
	Database    `mapstructure:"database"`
	SecretKey   []byte `mapstructure:"secret_key" env-required:"true"`
}

type HTTP_Server struct {
	Host         string `mapstructure:"host" env-default:"localhost:8080"`
	Timeout      int    `mapstructure:"timeout"`
	Idle_timeout int    `mapstructure:"idle_timeout"`
}
type Database struct {
	Host     string `mapstructuretructure:"host" env-default:"localhost"`
	Port     int    `mapstructure:"port" env-default:"5432"`
	User     string `mapstructure:"user" env-required:"true"`
	Password string `mapstructure:"pass" env-required:"true"`
	Name     string `mapstructure:"name" env-required:"true"`
}

func Loader() *Config {
	envConfig := os.Getenv("CONFIG_PATH")
	flagConfig := flag.String("configPath", "configs/dev", "config path")
	flag.Parse()

	var configPath string

	fmt.Println(*flagConfig)
	if envConfig == "" {
		slog.Info("Config path is not set in environment variables")
		configPath = *flagConfig
	} else {
		configPath = envConfig
	}
	slog.Info("Config path is now set: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist at path: %s", configPath)
	}

	var cfg Config

	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	return &cfg
}
