package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	DBDsn string
}

func Load() *Config {
	godotenv.Load()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASS", "postgres")
	name := getEnv("DB_NAME", "postgres")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)

	return &Config{
		Port:  getEnv("PORT", "8080"),
		DBDsn: getEnv("DB_DSN", dsn),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
