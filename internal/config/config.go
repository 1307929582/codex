package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort   string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	JWTSecret    string

	// OAuth configuration
	LinuxDoClientID     string
	LinuxDoClientSecret string
	LinuxDoRedirectURL  string
	FrontendURL         string
	DefaultBalance      float64
}

var AppConfig *Config

func Load() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		ServerPort:   getEnv("SERVER_PORT", "12322"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5433"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", ""),
		DBName:       getEnv("DB_NAME", "codex_gateway"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		JWTSecret:    getEnv("JWT_SECRET", ""),

		// OAuth configuration
		LinuxDoClientID:     getEnv("LINUXDO_CLIENT_ID", ""),
		LinuxDoClientSecret: getEnv("LINUXDO_CLIENT_SECRET", ""),
		LinuxDoRedirectURL:  getEnv("LINUXDO_REDIRECT_URL", "https://codex.zenscaleai.com/api/auth/linuxdo/callback"),
		FrontendURL:         getEnv("FRONTEND_URL", "https://codex.zenscaleai.com"),
		DefaultBalance:      getEnvFloat("DEFAULT_BALANCE", 0),
	}

	if AppConfig.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	if len(AppConfig.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var result float64
		if _, err := fmt.Sscanf(value, "%f", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
