package handlers

import (
	"net/http"
	"strconv"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
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
	query := database.DB.Where("user_id = ?", user.ID).Order("created_at desc").Limit(pageSize).Offset((page - 1) * pageSize)

	if model := c.Query("model"); model != "" {
		query = query.Where("model = ?", model)
	}

	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}

	var total int64
	database.DB.Model(&models.UsageLog{}).Where("user_id = ?", user.ID).Count(&total)

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
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	sevenDaysAgo := now.AddDate(0, 0, -7)

	var todayCost, monthCost, totalCost, sevenDaysCost float64
	var sevenDaysRequests, sevenDaysTokens int64

	database.DB.Model(&models.UsageLog{}).Where("user_id = ? AND created_at >= ?", user.ID, startOfDay).Select("COALESCE(SUM(cost), 0)").Scan(&todayCost)
	database.DB.Model(&models.UsageLog{}).Where("user_id = ? AND created_at >= ?", user.ID, startOfMonth).Select("COALESCE(SUM(cost), 0)").Scan(&monthCost)
	database.DB.Model(&models.UsageLog{}).Where("user_id = ?", user.ID).Select("COALESCE(SUM(cost), 0)").Scan(&totalCost)

	database.DB.Model(&models.UsageLog{}).Where("user_id = ? AND created_at >= ?", user.ID, sevenDaysAgo).Select("COALESCE(SUM(cost), 0)").Scan(&sevenDaysCost)
	database.DB.Model(&models.UsageLog{}).Where("user_id = ? AND created_at >= ?", user.ID, sevenDaysAgo).Count(&sevenDaysRequests)
	database.DB.Model(&models.UsageLog).Where("user_id = ? AND created_at >= ?", user.ID, sevenDaysAgo).Select("COALESCE(SUM(total_tokens), 0)").Scan(&sevenDaysTokens)

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
	now := time.Now()

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
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
