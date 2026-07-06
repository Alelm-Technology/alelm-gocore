package config

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	DBDsn string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env file not loaded: %v", err)
	}

	host := GetEnv("DB_HOST", "localhost")
	port := GetEnv("DB_PORT", "5432")
	user := GetEnv("DB_USER", "postgres")
	pass := GetEnv("DB_PASS", "postgres")
	name := GetEnv("DB_NAME", "postgres")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(user), url.PathEscape(pass), host, port, name)

	return &Config{
		Port:  GetEnv("PORT", "8080"),
		DBDsn: GetEnv("DB_DSN", dsn),
	}
}

func GetEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
