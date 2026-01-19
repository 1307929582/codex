package database

import (
	"codex-gateway/internal/models"
	"log"
)

// SeedCodexUpstreams seeds default Codex upstream configurations
func SeedCodexUpstreams() error {
	var count int64
	DB.Model(&models.CodexUpstream{}).Count(&count)
	if count > 0 {
		log.Println("Codex upstreams already seeded, skipping")
		return nil
	}

	// Default upstream configuration (to be configured by admin)
	defaultUpstream := models.CodexUpstream{
		Name:        "Default Codex Upstream",
		BaseURL:     "https://api.openai.com/v1",
		APIKey:      "YOUR_API_KEY_HERE", // Admin needs to configure this
		Priority:    0,
		Status:      "disabled", // Disabled by default until configured
		Weight:      1,
		MaxRetries:  3,
		Timeout:     120,
		HealthCheck: "/health",
	}

	if err := DB.Create(&defaultUpstream).Error; err != nil {
		log.Printf("Failed to seed Codex upstreams: %v", err)
		return err
	}

	log.Println("Codex upstreams seeded successfully")
	return nil
}
