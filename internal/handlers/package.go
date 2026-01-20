package handlers

import (
	"net/http"
	"strconv"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminListPackages lists all packages
func AdminListPackages(c *gin.Context) {
	var packages []models.Package
	if err := database.DB.Order("sort_order ASC, id ASC").Find(&packages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch packages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"packages": packages})
}

// AdminCreatePackage creates a new package
func AdminCreatePackage(c *gin.Context) {
	var req struct {
		Name         string  `json:"name" binding:"required"`
		Description  string  `json:"description"`
		Price        float64 `json:"price" binding:"required,gt=0"`
		DurationDays int     `json:"duration_days" binding:"required,gt=0"`
		DailyLimit   float64 `json:"daily_limit" binding:"required,gt=0"`
		SortOrder    int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	pkg := models.Package{
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		DurationDays: req.DurationDays,
		DailyLimit:   req.DailyLimit,
		Status:       "active",
		SortOrder:    req.SortOrder,
	}

	if err := database.DB.Create(&pkg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create package"})
		return
	}

	c.JSON(http.StatusOK, pkg)
}

// AdminUpdatePackage updates a package
func AdminUpdatePackage(c *gin.Context) {
	id := c.Param("id")

	var pkg models.Package
	if err := database.DB.First(&pkg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	var req struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		Price        float64 `json:"price"`
		DurationDays int     `json:"duration_days"`
		DailyLimit   float64 `json:"daily_limit"`
		SortOrder    int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Name != "" {
		pkg.Name = req.Name
	}
	pkg.Description = req.Description
	if req.Price > 0 {
		pkg.Price = req.Price
	}
	if req.DurationDays > 0 {
		pkg.DurationDays = req.DurationDays
	}
	if req.DailyLimit > 0 {
		pkg.DailyLimit = req.DailyLimit
	}
	pkg.SortOrder = req.SortOrder

	if err := database.DB.Save(&pkg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update package"})
		return
	}

	c.JSON(http.StatusOK, pkg)
}

// AdminDeletePackage deletes a package
func AdminDeletePackage(c *gin.Context) {
	id := c.Param("id")

	var pkg models.Package
	if err := database.DB.First(&pkg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	if err := database.DB.Delete(&pkg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete package"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "package deleted"})
}

// AdminUpdatePackageStatus updates package status
func AdminUpdatePackageStatus(c *gin.Context) {
	id := c.Param("id")

	var pkg models.Package
	if err := database.DB.First(&pkg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	pkg.Status = req.Status

	if err := database.DB.Save(&pkg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}

	c.JSON(http.StatusOK, pkg)
}

// ListPackages lists active packages for users
func ListPackages(c *gin.Context) {
	var packages []models.Package
	if err := database.DB.Where("status = ?", "active").Order("sort_order ASC, id ASC").Find(&packages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch packages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"packages": packages})
}

// GetUserPackages gets user's packages
func GetUserPackages(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.UserPackage{}).Where("user_id = ?", user.ID).Count(&total)

	var userPackages []models.UserPackage
	if err := database.DB.Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&userPackages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch packages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"packages": userPackages,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetUserDailyUsage gets user's daily usage
func GetUserDailyUsage(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	// Get today's usage
	var todayUsage models.DailyUsage
	today := database.GetToday()
	database.DB.Where("user_id = ? AND date = ?", user.ID, today).First(&todayUsage)

	// Get active package
	var activePackage *models.UserPackage
	database.DB.Where("user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
		user.ID, "active", today, today).
		Order("end_date ASC").
		First(&activePackage)

	response := gin.H{
		"date":        today,
		"used_amount": todayUsage.UsedAmount,
	}

	if activePackage != nil {
		response["package"] = activePackage
		response["daily_limit"] = activePackage.DailyLimit
		response["remaining"] = activePackage.DailyLimit - todayUsage.UsedAmount
	}

	c.JSON(http.StatusOK, response)
}
