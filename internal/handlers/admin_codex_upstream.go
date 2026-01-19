package handlers

import (
	"net/http"
	"strconv"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"
	upstreamSelector "codex-gateway/internal/upstream"

	"github.com/gin-gonic/gin"
)

// AdminListCodexUpstreams lists all Codex upstreams
func AdminListCodexUpstreams(c *gin.Context) {
	var upstreams []models.CodexUpstream

	if err := database.DB.Order("priority ASC, id ASC").Find(&upstreams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch upstreams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"upstreams": upstreams})
}

// AdminGetCodexUpstream gets a single upstream by ID
func AdminGetCodexUpstream(c *gin.Context) {
	id := c.Param("id")

	var upstream models.CodexUpstream
	if err := database.DB.Where("id = ?", id).First(&upstream).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upstream not found"})
		return
	}

	c.JSON(http.StatusOK, upstream)
}

// AdminCreateCodexUpstream creates a new upstream
func AdminCreateCodexUpstream(c *gin.Context) {
	var req models.CodexUpstream

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Set defaults
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Weight == 0 {
		req.Weight = 1
	}
	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}
	if req.Timeout == 0 {
		req.Timeout = 120
	}

	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upstream"})
		return
	}

	// Refresh upstream selector
	upstreamSelector.GetSelector().RefreshUpstreams()

	c.JSON(http.StatusCreated, req)
}

// AdminUpdateCodexUpstream updates an existing upstream
func AdminUpdateCodexUpstream(c *gin.Context) {
	id := c.Param("id")

	var upstream models.CodexUpstream
	if err := database.DB.Where("id = ?", id).First(&upstream).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "upstream not found"})
		return
	}

	var req models.CodexUpstream
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Update fields
	upstream.Name = req.Name
	upstream.BaseURL = req.BaseURL
	upstream.APIKey = req.APIKey
	upstream.Priority = req.Priority
	upstream.Status = req.Status
	upstream.Weight = req.Weight
	upstream.MaxRetries = req.MaxRetries
	upstream.Timeout = req.Timeout
	upstream.HealthCheck = req.HealthCheck

	if err := database.DB.Save(&upstream).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update upstream"})
		return
	}

	// Refresh upstream selector
	upstreamSelector.GetSelector().RefreshUpstreams()

	c.JSON(http.StatusOK, upstream)
}

// AdminDeleteCodexUpstream deletes an upstream
func AdminDeleteCodexUpstream(c *gin.Context) {
	id := c.Param("id")

	result := database.DB.Delete(&models.CodexUpstream{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete upstream"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "upstream not found"})
		return
	}

	// Refresh upstream selector
	upstreamSelector.GetSelector().RefreshUpstreams()

	c.JSON(http.StatusOK, gin.H{"message": "upstream deleted successfully"})
}

// AdminUpdateCodexUpstreamStatus updates upstream status
func AdminUpdateCodexUpstreamStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required,oneof=active disabled unhealthy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upstream ID"})
		return
	}

	result := database.DB.Model(&models.CodexUpstream{}).
		Where("id = ?", idInt).
		Update("status", req.Status)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "upstream not found"})
		return
	}

	// Refresh upstream selector
	upstreamSelector.GetSelector().RefreshUpstreams()

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}
