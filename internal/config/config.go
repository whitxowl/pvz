package config

import (
	"fmt"
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Env           string `env:"ENV" envDefault:"local"`
	HTTPServer    `env-prefix:"HTTP_"`
	StorageConfig `env-prefix:"DB_"`
	JWTConfig     `env-prefix:"JWT_"`
}

type HTTPServer struct {
	Address     string        `env:"ADDRESS" envDefault:"localhost:8080"`
	Timeout     time.Duration `env:"TIMEOUT" envDefault:"5s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
}

type StorageConfig struct {
	Host           string `env:"HOST" env-required:"true"`
	Port           string `env:"PORT" env-required:"true"`
	User           string `env:"USER" env-required:"true"`
	Password       string `env:"PASSWORD" env-required:"true"`
	Database       string `env:"NAME" env-required:"true"`
	SSLMode        string `env:"SSLMODE" env-default:"disable"`
	MaxConnections int    `env:"MAX_CONNECTIONS" env-default:"10"`
}

type JWTConfig struct {
	SecretKey            string        `env:"SECRET_KEY" env-required:"true"`
	AccessTokenDuration  time.Duration `env:"ACCESS_TOKEN_DURATION" env-default:"15m"`
	RefreshTokenDuration time.Duration `env:"REFRESH_TOKEN_DURATION" env-default:"720h"`
}

func (s *StorageConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		s.User,
		s.Password,
		s.Host,
		s.Port,
		s.Database,
		s.SSLMode)
}

func MustLoadConfig() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatal("Failed to read config: ", err)
	}

	return &cfg
}
