package database

import (
	"codex-gateway/internal/models"
	"log"
)

// SeedCodexPricing seeds Codex model pricing based on sub2api reference
func SeedCodexPricing() error {
	// Codex model pricing (per 1K tokens)
	// Exact pricing from sub2api's model_prices_and_context_window.json
	// No markup applied - using cost price
	codexModels := []models.ModelPricing{
		{
			ModelName:           "gpt-5.1-codex",
			InputPricePer1k:     0.00125,  // 1.25e-06 per token
			OutputPricePer1k:    0.01,     // 1e-05 per token
			CacheReadPricePer1k: 0.000125, // 1.25e-07 per token
			CacheCreationPricePer1k: 0.000125,
			MarkupMultiplier:    1.0,
		},
		{
			ModelName:           "gpt-5.1-codex-mini",
			InputPricePer1k:     0.00025,  // 2.5e-07 per token
			OutputPricePer1k:    0.002,    // 2e-06 per token
			CacheReadPricePer1k: 0.000025, // 2.5e-08 per token
			CacheCreationPricePer1k: 0.000025,
			MarkupMultiplier:    1.0,
		},
		{
			ModelName:           "gpt-5.1-codex-max",
			InputPricePer1k:     0.00125,  // Same as standard codex
			OutputPricePer1k:    0.01,
			CacheReadPricePer1k: 0.000125,
			CacheCreationPricePer1k: 0.000125,
			MarkupMultiplier:    1.0,
		},
		{
			ModelName:           "gpt-5.2-codex",
			InputPricePer1k:     0.00175,  // 1.75e-06 per token (gpt-5.2 pricing)
			OutputPricePer1k:    0.014,    // 1.4e-05 per token
			CacheReadPricePer1k: 0.000175, // 1.75e-07 per token
			CacheCreationPricePer1k: 0.000175,
			MarkupMultiplier:    1.0,
		},
		{
			ModelName:           "gpt-5.1",
			InputPricePer1k:     0.00125,
			OutputPricePer1k:    0.01,
			CacheReadPricePer1k: 0.000125,
			CacheCreationPricePer1k: 0.000125,
			MarkupMultiplier:    1.0,
		},
		{
			ModelName:           "gpt-5.2",
			InputPricePer1k:     0.00175,
			OutputPricePer1k:    0.014,
			CacheReadPricePer1k: 0.000175,
			CacheCreationPricePer1k: 0.000175,
			MarkupMultiplier:    1.0,
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
			existing.CacheReadPricePer1k = pricing.CacheReadPricePer1k
			existing.CacheCreationPricePer1k = pricing.CacheCreationPricePer1k
			existing.MarkupMultiplier = pricing.MarkupMultiplier
			if err := DB.Save(&existing).Error; err != nil {
				log.Printf("Failed to update pricing for %s: %v", pricing.ModelName, err)
				return err
			}
			log.Printf("Updated pricing for model: %s (input=$%.6f, output=$%.6f, cache_read=$%.6f, cache_create=$%.6f)",
				pricing.ModelName, pricing.InputPricePer1k, pricing.OutputPricePer1k, pricing.CacheReadPricePer1k, pricing.CacheCreationPricePer1k)
		}
	}

	log.Println("Codex pricing seeding completed")
	return nil
}
