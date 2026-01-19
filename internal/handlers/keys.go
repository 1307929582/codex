package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"codex-gateway/internal/database"
	"codex-gateway/internal/middleware"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

type CreateKeyRequest struct {
	Name       string   `json:"name" binding:"required"`
	QuotaLimit *float64 `json:"quota_limit"`
}

func ListAPIKeys(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	var keys []models.APIKey

	if err := database.DB.Where("user_id = ?", user.ID).Find(&keys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch keys"})
		return
	}

	c.JSON(http.StatusOK, keys)
}

func CreateAPIKey(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	var req CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
		return
	}
	rawKey := "sk-" + hex.EncodeToString(randomBytes)

	keyHash := middleware.HashAPIKey(rawKey)

	apiKey := models.APIKey{
		UserID:     user.ID,
		KeyHash:    keyHash,
		KeyPrefix:  rawKey[:7],
		Name:       req.Name,
		QuotaLimit: req.QuotaLimit,
		Status:     "active",
	}

	if err := database.DB.Create(&apiKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   apiKey.ID,
		"key":  rawKey,
		"name": apiKey.Name,
	})
}

func DeleteAPIKey(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	id := c.Param("id")

	if err := database.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&models.APIKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key deleted"})
}

func UpdateAPIKeyStatus(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required,oneof=active disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&models.APIKey{}).Where("id = ? AND user_id = ?", id, user.ID).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}
