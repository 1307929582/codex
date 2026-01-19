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
			ModelName:            "gpt-5.1-codex",
			InputPricePer1k:      0.00138,  // $0.00138 per 1K tokens
			OutputPricePer1k:     0.011,    // $0.011 per 1K tokens
			CachedInputPricePer1k: 0.00069, // 50% discount for cached tokens
			MarkupMultiplier:     1.5,
		},
		{
			ModelName:            "gpt-5.1-codex-mini",
			InputPricePer1k:      0.000275, // $0.000275 per 1K tokens
			OutputPricePer1k:     0.0022,   // $0.0022 per 1K tokens
			CachedInputPricePer1k: 0.0001375, // 50% discount
			MarkupMultiplier:     1.5,
		},
		{
			ModelName:            "gpt-5.1-codex-max",
			InputPricePer1k:      0.00138,  // Same as standard codex
			OutputPricePer1k:     0.011,
			CachedInputPricePer1k: 0.00069,
			MarkupMultiplier:     1.5,
		},
		{
			ModelName:            "gpt-5.2-codex",
			InputPricePer1k:      0.00138,
			OutputPricePer1k:     0.011,
			CachedInputPricePer1k: 0.00069,
			MarkupMultiplier:     1.5,
		},
		{
			ModelName:            "gpt-5.1",
			InputPricePer1k:      0.00138,
			OutputPricePer1k:     0.011,
			CachedInputPricePer1k: 0.00069,
			MarkupMultiplier:     1.5,
		},
		{
			ModelName:            "gpt-5.2",
			InputPricePer1k:      0.00138,
			OutputPricePer1k:     0.011,
			CachedInputPricePer1k: 0.00069,
			MarkupMultiplier:     1.5,
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
			// Model exists, update it to ensure correct pricing
			existing.InputPricePer1k = pricing.InputPricePer1k
			existing.OutputPricePer1k = pricing.OutputPricePer1k
			existing.CachedInputPricePer1k = pricing.CachedInputPricePer1k
			existing.MarkupMultiplier = pricing.MarkupMultiplier
			if err := DB.Save(&existing).Error; err != nil {
				log.Printf("Failed to update pricing for %s: %v", pricing.ModelName, err)
				return err
			}
			log.Printf("Updated pricing for model: %s (input=$%.6f, output=$%.6f, cached=$%.6f)",
				pricing.ModelName, pricing.InputPricePer1k, pricing.OutputPricePer1k, pricing.CachedInputPricePer1k)
		}
	}

	log.Println("Codex pricing seeding completed")
	return nil
}
