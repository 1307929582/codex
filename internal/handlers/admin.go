package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminListUsers lists all users with pagination
func AdminListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	status := c.Query("status")

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.User{})

	if search != "" {
		query = query.Where("email LIKE ?", "%"+search+"%")
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var users []models.User
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// AdminGetUser gets a single user by ID
func AdminGetUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get user's API keys count
	var apiKeyCount int64
	database.DB.Model(&models.APIKey{}).Where("user_id = ?", userID).Count(&apiKeyCount)

	// Get user's total usage
	var totalUsage struct {
		TotalCost   float64
		TotalTokens int64
	}
	database.DB.Model(&models.UsageLog{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(cost), 0) as total_cost, COALESCE(SUM(total_tokens), 0) as total_tokens").
		Scan(&totalUsage)

	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"api_key_count": apiKeyCount,
		"total_cost":    totalUsage.TotalCost,
		"total_tokens":  totalUsage.TotalTokens,
	})
}

// AdminUpdateBalance updates a user's balance
func AdminUpdateBalance(c *gin.Context) {
	admin := c.MustGet("admin").(models.User)
	userID := c.Param("id")

	var req struct {
		Amount      float64 `json:"amount" binding:"required"`
		Description string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Update balance and create transaction
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.User{}).
			Where("id = ?", uid).
			Update("balance", gorm.Expr("balance + ?", req.Amount))

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("user not found")
		}

		// Create transaction record
		txn := models.Transaction{
			UserID:      uid,
			Amount:      req.Amount,
			Type:        "admin_adjustment",
			Description: fmt.Sprintf("Admin adjustment by %s: %s", admin.Email, req.Description),
		}

		if err := tx.Create(&txn).Error; err != nil {
			return err
		}

		// Log admin action
		log := models.AdminLog{
			AdminID:   admin.ID,
			Action:    "update_balance",
			Target:    userID,
			Details:   fmt.Sprintf("Amount: %.6f, Description: %s", req.Amount, req.Description),
			IPAddress: c.ClientIP(),
		}

		return tx.Create(&log).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "balance updated successfully"})
}

// AdminUpdateUserStatus updates a user's status
func AdminUpdateUserStatus(c *gin.Context) {
	admin := c.MustGet("admin").(models.User)
	userID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required,oneof=active suspended banned"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.User{}).
			Where("id = ?", uid).
			Update("status", req.Status)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("user not found")
		}

		// Log admin action
		log := models.AdminLog{
			AdminID:   admin.ID,
			Action:    "update_user_status",
			Target:    userID,
			Details:   fmt.Sprintf("New status: %s", req.Status),
			IPAddress: c.ClientIP(),
		}

		return tx.Create(&log).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user status updated successfully"})
}

// AdminGetSettings gets system settings
func AdminGetSettings(c *gin.Context) {
	var settings models.SystemSettings

	if err := database.DB.First(&settings).Error; err != nil {
		// If no settings exist, return defaults
		settings = models.SystemSettings{
			Announcement:        "",
			DefaultBalance:      0,
			MinRechargeAmount:   10,
			RegistrationEnabled: true,
		}
	}

	c.JSON(http.StatusOK, settings)
}

// AdminUpdateSettings updates system settings
func AdminUpdateSettings(c *gin.Context) {
	admin := c.MustGet("admin").(models.User)

	var req models.SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var settings models.SystemSettings
		result := tx.First(&settings)

		if result.Error != nil {
			// Create new settings
			req.ID = 1
			if err := tx.Create(&req).Error; err != nil {
				return err
			}
		} else {
			// Update existing settings
			if err := tx.Model(&settings).Updates(req).Error; err != nil {
				return err
			}
		}

		// Log admin action
		log := models.AdminLog{
			AdminID:   admin.ID,
			Action:    "update_settings",
			Target:    "system",
			Details:   fmt.Sprintf("Updated system settings"),
			IPAddress: c.ClientIP(),
		}

		return tx.Create(&log).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings updated successfully"})
}

// AdminGetOverview gets system overview statistics
func AdminGetOverview(c *gin.Context) {
	var stats struct {
		TotalUsers      int64
		ActiveUsers     int64
		TotalRevenue    float64
		TotalCost       float64
		TotalAPIKeys    int64
		TodayRequests   int64
		TodayRevenue    float64
	}

	// Total users
	database.DB.Model(&models.User{}).Count(&stats.TotalUsers)

	// Active users
	database.DB.Model(&models.User{}).Where("status = ?", "active").Count(&stats.ActiveUsers)

	// Total revenue (deposits)
	database.DB.Model(&models.Transaction{}).
		Where("type IN ?", []string{"deposit", "admin_adjustment"}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalRevenue)

	// Total cost (usage)
	database.DB.Model(&models.UsageLog{}).
		Select("COALESCE(SUM(cost), 0)").
		Scan(&stats.TotalCost)

	// Total API keys
	database.DB.Model(&models.APIKey{}).Count(&stats.TotalAPIKeys)

	// Today's requests
	database.DB.Model(&models.UsageLog{}).
		Where("DATE(created_at) = CURRENT_DATE").
		Count(&stats.TodayRequests)

	// Today's revenue
	database.DB.Model(&models.UsageLog{}).
		Where("DATE(created_at) = CURRENT_DATE").
		Select("COALESCE(SUM(cost), 0)").
		Scan(&stats.TodayRevenue)

	c.JSON(http.StatusOK, stats)
}

// AdminGetLogs gets admin operation logs
func AdminGetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.AdminLog{}).Count(&total)

	var logs []models.AdminLog
	if err := database.DB.Preload("Admin").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}
