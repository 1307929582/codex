package database

import (
	"fmt"
	"log"
	"time"

	"codex-gateway/internal/config"
	"codex-gateway/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() error {
	cfg := config.AppConfig
	if cfg.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required but not set")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	var err error
	// Use Error level logging in production to reduce overhead
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for high concurrency
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
	// Increased from 10 to 25 to reduce connection churn under spiky workloads
	sqlDB.SetMaxIdleConns(25)

	// SetMaxOpenConns sets the maximum number of open connections to the database
	// This allows up to 100 concurrent database operations per backend instance
	// Note: If running multiple backend replicas, ensure total connections < PostgreSQL max_connections
	// Example: 3 replicas × 100 = 300 connections (PostgreSQL max_connections should be >= 350)
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
	// This helps prevent issues with stale connections
	sqlDB.SetConnMaxLifetime(time.Hour)

	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle
	// Connections idle longer than this will be closed to free resources
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	log.Println("Database connected successfully with optimized connection pool configured")
	return nil
}

func AutoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.APIKey{},
		&models.ModelPricing{},
		&models.UsageLog{},
		&models.Transaction{},
		&models.SystemSettings{},
		&models.AdminLog{},
		&models.CodexUpstream{},
		&models.Package{},
		&models.UserPackage{},
		&models.DailyUsage{},
		&models.PaymentOrder{},
	)
}

func SeedDefaultPricing() error {
	var count int64
	DB.Model(&models.ModelPricing{}).Count(&count)
	if count > 0 {
		return nil
	}

	defaultPricing := []models.ModelPricing{
		{
			ModelName:        "gpt-4",
			InputPricePer1k:  0.03,
			OutputPricePer1k: 0.06,
			MarkupMultiplier: 1.5,
		},
		{
			ModelName:        "gpt-4-turbo",
			InputPricePer1k:  0.01,
			OutputPricePer1k: 0.03,
			MarkupMultiplier: 1.5,
		},
		{
			ModelName:        "gpt-3.5-turbo",
			InputPricePer1k:  0.0005,
			OutputPricePer1k: 0.0015,
			MarkupMultiplier: 1.5,
		},
	}

	return DB.Create(&defaultPricing).Error
}
