package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetUsageLogs(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var logs []models.UsageLog
	query := database.DB.Model(&models.UsageLog{}).Where("user_id = ?", user.ID)

	if model := c.Query("model"); model != "" {
		query = query.Where("model = ?", model)
	}

	startTime, endTime, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at < ?", *endTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count logs"})
		return
	}

	if err := query.Order("created_at desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       logs,
		"page":       page,
		"page_size":  pageSize,
		"total":      total,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func GetUsageStats(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	shanghaiTZ, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(shanghaiTZ)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, shanghaiTZ).UTC()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, shanghaiTZ).UTC()
	sevenDaysAgo := now.AddDate(0, 0, -7).UTC()

	var todayCost, monthCost, totalCost, sevenDaysCost float64
	var sevenDaysRequests, sevenDaysTokens int64

	database.DB.Model(&models.UsageLog{}).Where("user_id = ? AND created_at >= ?", user.ID, startOfDay).Select("COALESCE(SUM(cost), 0)").Scan(&todayCost)
	database.DB.Model(&models.UsageLog{}).Where("user_id = ? AND created_at >= ?", user.ID, startOfMonth).Select("COALESCE(SUM(cost), 0)").Scan(&monthCost)
	database.DB.Model(&models.UsageLog{}).Where("user_id = ?", user.ID).Select("COALESCE(SUM(cost), 0)").Scan(&totalCost)

	database.DB.Model(&models.UsageLog{}).Where("user_id = ? AND created_at >= ?", user.ID, sevenDaysAgo).Select("COALESCE(SUM(cost), 0)").Scan(&sevenDaysCost)
	database.DB.Model(&models.UsageLog{}).Where("user_id = ? AND created_at >= ?", user.ID, sevenDaysAgo).Count(&sevenDaysRequests)
	database.DB.Model(&models.UsageLog{}).Where("user_id = ? AND created_at >= ?", user.ID, sevenDaysAgo).Select("COALESCE(SUM(total_tokens), 0)").Scan(&sevenDaysTokens)

	c.JSON(http.StatusOK, gin.H{
		"today_cost":          todayCost,
		"month_cost":          monthCost,
		"total_cost":          totalCost,
		"seven_days_cost":     sevenDaysCost,
		"seven_days_requests": sevenDaysRequests,
		"seven_days_tokens":   sevenDaysTokens,
	})
}

func GetBalance(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	c.JSON(http.StatusOK, gin.H{
		"balance":  user.Balance,
		"currency": "USD",
	})
}

func GetTransactions(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	var transactions []models.Transaction

	if err := database.DB.Where("user_id = ?", user.ID).Order("created_at desc").Limit(50).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func GetDailyTrend(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	trendType := c.DefaultQuery("type", "cost")

	type DailyData struct {
		Date  string  `json:"date"`
		Value float64 `json:"value"`
	}

	var results []DailyData
	shanghaiTZ, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(shanghaiTZ)

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, shanghaiTZ).UTC()
		endOfDay := startOfDay.Add(24 * time.Hour)
		dateStr := date.Format("01-02")

		var value float64
		switch trendType {
		case "cost":
			database.DB.Model(&models.UsageLog{}).
				Where("user_id = ? AND created_at >= ? AND created_at < ?", user.ID, startOfDay, endOfDay).
				Select("COALESCE(SUM(cost), 0)").
				Scan(&value)
		case "requests":
			var count int64
			database.DB.Model(&models.UsageLog{}).
				Where("user_id = ? AND created_at >= ? AND created_at < ?", user.ID, startOfDay, endOfDay).
				Count(&count)
			value = float64(count)
		case "tokens":
			database.DB.Model(&models.UsageLog{}).
				Where("user_id = ? AND created_at >= ? AND created_at < ?", user.ID, startOfDay, endOfDay).
				Select("COALESCE(SUM(total_tokens), 0)").
				Scan(&value)
		}

		results = append(results, DailyData{
			Date:  dateStr,
			Value: value,
		})
	}

	c.JSON(http.StatusOK, results)
}

// AdminGetUsageLogs gets usage logs with filters for admin
func AdminGetUsageLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&models.UsageLog{})

	startTime, endTime, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at < ?", *endTime)
	}

	if model := strings.TrimSpace(c.Query("model")); model != "" {
		query = query.Where("model = ?", model)
	}

	if statusCodeStr := strings.TrimSpace(c.Query("status_code")); statusCodeStr != "" {
		statusCode, err := strconv.Atoi(statusCodeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status_code"})
			return
		}
		query = query.Where("status_code = ?", statusCode)
	}

	if apiKeyIDStr := strings.TrimSpace(c.Query("api_key_id")); apiKeyIDStr != "" {
		apiKeyID, err := strconv.Atoi(apiKeyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid api_key_id"})
			return
		}
		query = query.Where("api_key_id = ?", apiKeyID)
	}

	if userIDStr := strings.TrimSpace(c.Query("user_id")); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		query = query.Where("usage_logs.user_id = ?", userID)
	}

	if userFilter := strings.TrimSpace(c.Query("user")); userFilter != "" {
		if userID, err := uuid.Parse(userFilter); err == nil {
			query = query.Where("usage_logs.user_id = ?", userID)
		} else {
			like := "%" + userFilter + "%"
			query = query.Joins("JOIN users ON users.id = usage_logs.user_id").
				Where("users.username ILIKE ? OR users.oauth_id ILIKE ?", like, like)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count logs"})
		return
	}

	var logs []models.UsageLog
	if err := query.Preload("User").
		Order("created_at desc").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}

	type AdminUsageLog struct {
		RequestID    string  `json:"request_id"`
		UserID       string  `json:"user_id"`
		Username     string  `json:"username"`
		LinuxDoID    string  `json:"linuxdo_id"`
		APIKeyID     uint    `json:"api_key_id"`
		Model        string  `json:"model"`
		InputTokens  int     `json:"input_tokens"`
		OutputTokens int     `json:"output_tokens"`
		CachedTokens int     `json:"cached_tokens"`
		TotalTokens  int     `json:"total_tokens"`
		Cost         float64 `json:"cost"`
		LatencyMs    int     `json:"latency_ms"`
		StatusCode   int     `json:"status_code"`
		CreatedAt    string  `json:"created_at"`
	}

	response := make([]AdminUsageLog, 0, len(logs))
	for _, log := range logs {
		linuxdoID := ""
		if log.User.OAuthProvider == "linuxdo" {
			linuxdoID = log.User.OAuthID
		}
		response = append(response, AdminUsageLog{
			RequestID:    log.RequestID.String(),
			UserID:       log.UserID.String(),
			Username:     log.User.Username,
			LinuxDoID:    linuxdoID,
			APIKeyID:     log.APIKeyID,
			Model:        log.Model,
			InputTokens:  log.InputTokens,
			OutputTokens: log.OutputTokens,
			CachedTokens: log.CachedTokens,
			TotalTokens:  log.TotalTokens,
			Cost:         log.Cost,
			LatencyMs:    log.LatencyMs,
			StatusCode:   log.StatusCode,
			CreatedAt:    log.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": response,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func parseDateRange(c *gin.Context) (*time.Time, *time.Time, error) {
	startStr := strings.TrimSpace(c.Query("start_date"))
	endStr := strings.TrimSpace(c.Query("end_date"))
	if startStr == "" && endStr == "" {
		return nil, nil, nil
	}

	var startTime *time.Time
	var endTime *time.Time

	if startStr != "" {
		startDate, err := time.ParseInLocation("2006-01-02", startStr, database.AsiaShanghai)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid start_date")
		}
		startUTC := startDate.UTC()
		startTime = &startUTC
	}

	if endStr != "" {
		endDate, err := time.ParseInLocation("2006-01-02", endStr, database.AsiaShanghai)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid end_date")
		}
		endExclusive := endDate.AddDate(0, 0, 1).UTC()
		endTime = &endExclusive
	}

	if startTime != nil && endTime != nil && endTime.Before(*startTime) {
		return nil, nil, fmt.Errorf("end_date must be after start_date")
	}

	return startTime, endTime, nil
}
