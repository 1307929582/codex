package database

import (
	"fmt"
	"log"

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
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
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
			ModelName:          "gpt-4",
			InputPricePer1k:    0.03,
			OutputPricePer1k:   0.06,
			MarkupMultiplier:   1.5,
		},
		{
			ModelName:          "gpt-4-turbo",
			InputPricePer1k:    0.01,
			OutputPricePer1k:   0.03,
			MarkupMultiplier:   1.5,
		},
		{
			ModelName:          "gpt-3.5-turbo",
			InputPricePer1k:    0.0005,
			OutputPricePer1k:   0.0015,
			MarkupMultiplier:   1.5,
		},
	}

	return DB.Create(&defaultPricing).Error
}
