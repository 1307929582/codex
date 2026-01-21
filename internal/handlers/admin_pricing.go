package handlers

import (
	"net/http"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminListPricing lists all model pricing
func AdminListPricing(c *gin.Context) {
	var pricing []models.ModelPricing
	if err := database.DB.Order("model_name ASC").Find(&pricing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pricing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pricing": pricing})
}

// AdminUpdatePricing updates a model's pricing
func AdminUpdatePricing(c *gin.Context) {
	id := c.Param("id")

	var pricing models.ModelPricing
	if err := database.DB.First(&pricing, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pricing not found"})
		return
	}

	var req struct {
		InputPricePer1k         *float64 `json:"input_price_per_1k"`
		OutputPricePer1k        *float64 `json:"output_price_per_1k"`
		CacheReadPricePer1k     *float64 `json:"cache_read_price_per_1k"`
		CacheCreationPricePer1k *float64 `json:"cache_creation_price_per_1k"`
		MarkupMultiplier        *float64 `json:"markup_multiplier"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Update only provided fields
	if req.InputPricePer1k != nil {
		pricing.InputPricePer1k = *req.InputPricePer1k
	}
	if req.OutputPricePer1k != nil {
		pricing.OutputPricePer1k = *req.OutputPricePer1k
	}
	if req.CacheReadPricePer1k != nil {
		pricing.CacheReadPricePer1k = *req.CacheReadPricePer1k
	}
	if req.CacheCreationPricePer1k != nil {
		pricing.CacheCreationPricePer1k = *req.CacheCreationPricePer1k
	}
	if req.MarkupMultiplier != nil {
		pricing.MarkupMultiplier = *req.MarkupMultiplier
	}

	if err := database.DB.Save(&pricing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update pricing"})
		return
	}

	c.JSON(http.StatusOK, pricing)
}

// AdminBatchUpdateMarkup updates markup multiplier for all models
func AdminBatchUpdateMarkup(c *gin.Context) {
	var req struct {
		MarkupMultiplier float64 `json:"markup_multiplier" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Update all models - use Where clause to ensure all records are updated
	result := database.DB.Model(&models.ModelPricing{}).
		Where("id > ?", 0).
		Updates(map[string]interface{}{
			"markup_multiplier": req.MarkupMultiplier,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update markup"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "markup updated successfully",
		"updated_count":     result.RowsAffected,
		"markup_multiplier": req.MarkupMultiplier,
	})
}

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
