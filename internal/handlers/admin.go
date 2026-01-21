package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"
	"codex-gateway/internal/pricing"
	"codex-gateway/internal/upstream"

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
		query = query.Where("email LIKE ? OR username LIKE ?", "%"+search+"%", "%"+search+"%")
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

	// Get active packages for each user
	today := database.GetToday()
	type UserWithPackage struct {
		models.User
		ActivePackage *models.UserPackage `json:"active_package"`
	}

	var usersWithPackages []UserWithPackage
	for _, user := range users {
		var activePackage models.UserPackage
		err := database.DB.Where("user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
			user.ID, "active", today, today).
			Order("end_date ASC").
			First(&activePackage).Error

		userWithPkg := UserWithPackage{
			User: user,
		}
		if err == nil {
			userWithPkg.ActivePackage = &activePackage
		}
		usersWithPackages = append(usersWithPackages, userWithPkg)
	}

	c.JSON(http.StatusOK, gin.H{
		"users": usersWithPackages,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
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
			Announcement:               "",
			DefaultBalance:             0,
			MinRechargeAmount:          10,
			EmailRegistrationEnabled:   true,
			LinuxDoRegistrationEnabled: true,
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

	if req.OpenAIBaseURL == "" {
		req.OpenAIBaseURL = "https://api.openai.com/v1"
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
			// Use map to update all fields including zero values (false for booleans)
			updates := map[string]interface{}{
				"announcement":                  req.Announcement,
				"default_balance":               req.DefaultBalance,
				"min_recharge_amount":           req.MinRechargeAmount,
				"email_registration_enabled":    req.EmailRegistrationEnabled,
				"linux_do_registration_enabled": req.LinuxDoRegistrationEnabled,
				"openai_api_key":                req.OpenAIAPIKey,
				"openai_base_url":               req.OpenAIBaseURL,
				"linuxdo_client_id":             req.LinuxDoClientID,
				"linuxdo_client_secret":         req.LinuxDoClientSecret,
				"linuxdo_enabled":               req.LinuxDoEnabled,
				"credit_enabled":                req.CreditEnabled,
				"credit_pid":                    req.CreditPID,
				"credit_key":                    req.CreditKey,
				"credit_notify_url":             req.CreditNotifyURL,
				"credit_return_url":             req.CreditReturnURL,
			}
			if err := tx.Model(&settings).Updates(updates).Error; err != nil {
				return err
			}
		}

		// Keep default upstream in sync with OpenAI settings when key is provided
		if req.OpenAIAPIKey != "" {
			var defaultUpstream models.CodexUpstream
			err := tx.Where("name = ?", "Default Codex Upstream").First(&defaultUpstream).Error
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				defaultUpstream = models.CodexUpstream{
					Name:        "Default Codex Upstream",
					BaseURL:     req.OpenAIBaseURL,
					APIKey:      req.OpenAIAPIKey,
					Priority:    0,
					Status:      "active",
					Weight:      1,
					MaxRetries:  3,
					Timeout:     120,
					HealthCheck: "/health",
				}
				if err := tx.Create(&defaultUpstream).Error; err != nil {
					return err
				}
			} else {
				defaultUpstream.BaseURL = req.OpenAIBaseURL
				defaultUpstream.APIKey = req.OpenAIAPIKey
				defaultUpstream.Status = "active"
				if defaultUpstream.Weight == 0 {
					defaultUpstream.Weight = 1
				}
				if defaultUpstream.MaxRetries == 0 {
					defaultUpstream.MaxRetries = 3
				}
				if defaultUpstream.Timeout == 0 {
					defaultUpstream.Timeout = 120
				}
				if defaultUpstream.HealthCheck == "" {
					defaultUpstream.HealthCheck = "/health"
				}
				if err := tx.Save(&defaultUpstream).Error; err != nil {
					return err
				}
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

	// Refresh upstream selector after settings update
	_ = upstream.GetSelector().RefreshUpstreams()
}

// AdminGetOverview gets system overview statistics
func AdminGetOverview(c *gin.Context) {
	var stats struct {
		TotalUsers    int64   `json:"total_users"`
		ActiveUsers   int64   `json:"active_users"`
		TotalTokens   int64   `json:"total_tokens"`
		TotalCost     float64 `json:"total_cost"`
		TotalAPIKeys  int64   `json:"total_api_keys"`
		TodayRequests int64   `json:"today_requests"`
		TodayRevenue  float64 `json:"today_revenue"`
	}

	// Total users
	database.DB.Model(&models.User{}).Count(&stats.TotalUsers)

	// Active users
	database.DB.Model(&models.User{}).Where("status = ?", "active").Count(&stats.ActiveUsers)

	// Total tokens
	database.DB.Model(&models.UsageLog{}).
		Select("COALESCE(SUM(total_tokens), 0)").
		Scan(&stats.TotalTokens)

	// Total cost (usage)
	database.DB.Model(&models.UsageLog{}).
		Select("COALESCE(SUM(cost), 0)").
		Scan(&stats.TotalCost)

	// Total API keys
	database.DB.Model(&models.APIKey{}).Count(&stats.TotalAPIKeys)

	// Calculate today's date range in UTC for Asia/Shanghai timezone
	// This avoids using timezone functions in WHERE clause which prevents index usage
	shanghaiTZ, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(shanghaiTZ)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, shanghaiTZ).UTC()
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Today's requests (using UTC range for index efficiency)
	database.DB.Model(&models.UsageLog{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Count(&stats.TodayRequests)

	// Today's revenue (using UTC range for index efficiency)
	database.DB.Model(&models.UsageLog{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Select("COALESCE(SUM(cost), 0)").
		Scan(&stats.TodayRevenue)

	c.JSON(http.StatusOK, stats)
}

// AdminGetUsageChart gets usage statistics for different time ranges
func AdminGetUsageChart(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h") // 24h, 7d, 30d

	type UsageData struct {
		Label string  `json:"label"`
		Cost  float64 `json:"cost"`
	}

	var usageData []UsageData
	shanghaiTZ, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(shanghaiTZ)

	switch timeRange {
	case "24h":
		// Last 24 hours - hourly data
		twentyFourHoursAgo := now.Add(-24 * time.Hour).UTC()
		database.DB.Model(&models.UsageLog{}).
			Select("TO_CHAR(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Shanghai', 'HH24:00') as label, COALESCE(SUM(cost), 0) as cost").
			Where("created_at >= ?", twentyFourHoursAgo).
			Group("TO_CHAR(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Shanghai', 'HH24:00')").
			Order("label").
			Scan(&usageData)

	case "7d":
		// Last 7 days - daily data
		sevenDaysAgo := now.AddDate(0, 0, -7).UTC()
		database.DB.Model(&models.UsageLog{}).
			Select("TO_CHAR(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Shanghai', 'MM-DD') as label, COALESCE(SUM(cost), 0) as cost").
			Where("created_at >= ?", sevenDaysAgo).
			Group("TO_CHAR(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Shanghai', 'MM-DD')").
			Order("label").
			Scan(&usageData)

	case "30d":
		// Last 30 days - daily data
		thirtyDaysAgo := now.AddDate(0, 0, -30).UTC()
		database.DB.Model(&models.UsageLog{}).
			Select("TO_CHAR(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Shanghai', 'MM-DD') as label, COALESCE(SUM(cost), 0) as cost").
			Where("created_at >= ?", thirtyDaysAgo).
			Group("TO_CHAR(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Shanghai', 'MM-DD')").
			Order("label").
			Scan(&usageData)
	}

	c.JSON(http.StatusOK, usageData)
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

// AdminGetPricingStatus gets pricing service status
func AdminGetPricingStatus(c *gin.Context) {
	pricingService := pricing.GetService()
	status := pricingService.GetStatus()
	c.JSON(http.StatusOK, status)
}
