package configs

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env            string `yaml:"env" env-default:"local"`
	HTTPServer     `yaml:"http_server"`
	SQLConfig      `yaml:"sql_config"`
	SteamAPIConfig `yaml:"steam_api"`
}

type HTTPServer struct {
	Address string        `yaml:"address" env-default:"localhost:8080"`
	Timeout time.Duration `yaml:"timeout" env-default:"4s"`
}

type SQLConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	Database string `yaml:"dbname" env-default:"postgres"`
}

type SteamAPIConfig struct {
	Key string `yaml:"key" env-required:"true"`
}

func MustLoad() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Print("No .env file found")
	}
	configPath := os.Getenv("CONFIG_PATH") //Путь задается из переменной окружения
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return cfg
}
