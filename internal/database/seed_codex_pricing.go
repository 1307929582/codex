package database

import (
	"codex-gateway/internal/models"
	"log"
)

// SeedCodexPricing seeds Codex model pricing based on sub2api reference
func SeedCodexPricing() error {
	// Codex model pricing (per 1K tokens)
	// Based on sub2api pricing with 1.5x markup
	codexModels := []models.ModelPricing{
		{
			ModelName:        "gpt-5.1-codex",
			InputPricePer1k:  0.00138,  // $0.00138 per 1K tokens
			OutputPricePer1k: 0.011,    // $0.011 per 1K tokens
			MarkupMultiplier: 1.5,
		},
		{
			ModelName:        "gpt-5.1-codex-mini",
			InputPricePer1k:  0.000275, // $0.000275 per 1K tokens
			OutputPricePer1k: 0.0022,   // $0.0022 per 1K tokens
			MarkupMultiplier: 1.5,
		},
		{
			ModelName:        "gpt-5.1-codex-max",
			InputPricePer1k:  0.00138,  // Same as standard codex
			OutputPricePer1k: 0.011,
			MarkupMultiplier: 1.5,
		},
		{
			ModelName:        "gpt-5.2-codex",
			InputPricePer1k:  0.00138,
			OutputPricePer1k: 0.011,
			MarkupMultiplier: 1.5,
		},
		{
			ModelName:        "gpt-5.1",
			InputPricePer1k:  0.00138,
			OutputPricePer1k: 0.011,
			MarkupMultiplier: 1.5,
		},
		{
			ModelName:        "gpt-5.2",
			InputPricePer1k:  0.00138,
			OutputPricePer1k: 0.011,
			MarkupMultiplier: 1.5,
		},
	}

	for _, pricing := range codexModels {
		// Check if model already exists
		var existing models.ModelPricing
		result := DB.Where("model_name = ?", pricing.ModelName).First(&existing)

		if result.Error != nil {
			// Model doesn't exist, create it
			if err := DB.Create(&pricing).Error; err != nil {
				log.Printf("Failed to seed pricing for %s: %v", pricing.ModelName, err)
				return err
			}
			log.Printf("Seeded pricing for model: %s", pricing.ModelName)
		} else {
			log.Printf("Pricing for %s already exists, skipping", pricing.ModelName)
		}
	}

	log.Println("Codex pricing seeding completed")
	return nil
}
