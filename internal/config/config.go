package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort      string
	OpenAIAPIKey    string
	OpenAIBaseURL   string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	JWTSecret       string
}

var AppConfig *Config

func Load() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:   getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "codex_gateway"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
	}

	if AppConfig.OpenAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
