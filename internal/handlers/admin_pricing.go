package handlers

import (
	"net/http"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminResetPricing resets all model pricing to correct values
func AdminResetPricing(c *gin.Context) {
	// Delete all existing pricing
	if err := database.DB.Where("model_name LIKE ?", "gpt-%").Delete(&models.ModelPricing{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete old pricing"})
		return
	}

	// Re-seed pricing
	if err := database.SeedCodexPricing(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to seed pricing"})
		return
	}

	// Get updated pricing to return
	var pricing []models.ModelPricing
	if err := database.DB.Where("model_name LIKE ?", "gpt-%").Find(&pricing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pricing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pricing reset successfully",
		"pricing": pricing,
	})
}
