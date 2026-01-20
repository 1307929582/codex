package handlers

import (
	"net/http"
	"strconv"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminListOrders lists all payment orders
func AdminListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	userID := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.PaymentOrder{})

	// Filter by status
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by user_id
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get orders with user info
	var orders []models.PaymentOrder
	err := query.
		Preload("User").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&orders).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response
	type OrderResponse struct {
		ID            string  `json:"id"`
		OrderNo       string  `json:"order_no"`
		UserID        string  `json:"user_id"`
		UserEmail     string  `json:"user_email"`
		Username      string  `json:"username"`
		PackageID     *uint   `json:"package_id"`
		Amount        float64 `json:"amount"`
		Status        string  `json:"status"`
		PaymentMethod string  `json:"payment_method"`
		TradeNo       string  `json:"trade_no"`
		CreatedAt     string  `json:"created_at"`
		PaidAt        *string `json:"paid_at"`
	}

	var response []OrderResponse
	for _, order := range orders {
		var paidAt *string
		if order.PaidAt != nil {
			paidAtStr := order.PaidAt.Format("2006-01-02 15:04:05")
			paidAt = &paidAtStr
		}

		response = append(response, OrderResponse{
			ID:            order.ID.String(),
			OrderNo:       order.OrderNo,
			UserID:        order.UserID.String(),
			UserEmail:     order.User.Email,
			Username:      order.User.Username,
			PackageID:     order.PackageID,
			Amount:        order.Amount,
			Status:        order.Status,
			PaymentMethod: order.PaymentMethod,
			TradeNo:       order.TradeNo,
			CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04:05"),
			PaidAt:        paidAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": response,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// AdminGetOrderStats gets order statistics
func AdminGetOrderStats(c *gin.Context) {
	var stats struct {
		TotalOrders    int64   `json:"total_orders"`
		PendingOrders  int64   `json:"pending_orders"`
		PaidOrders     int64   `json:"paid_orders"`
		FailedOrders   int64   `json:"failed_orders"`
		TotalRevenue   float64 `json:"total_revenue"`
		TodayRevenue   float64 `json:"today_revenue"`
		MonthRevenue   float64 `json:"month_revenue"`
	}

	// Total orders
	database.DB.Model(&models.PaymentOrder{}).Count(&stats.TotalOrders)

	// Pending orders
	database.DB.Model(&models.PaymentOrder{}).Where("status = ?", "pending").Count(&stats.PendingOrders)

	// Paid orders
	database.DB.Model(&models.PaymentOrder{}).Where("status = ?", "paid").Count(&stats.PaidOrders)

	// Failed orders
	database.DB.Model(&models.PaymentOrder{}).Where("status = ?", "failed").Count(&stats.FailedOrders)

	// Total revenue
	database.DB.Model(&models.PaymentOrder{}).
		Where("status = ?", "paid").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalRevenue)

	// Calculate today's date range in UTC
	// This avoids using DATE() function in WHERE clause which prevents index usage
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Today revenue (using UTC range for index efficiency)
	database.DB.Model(&models.PaymentOrder{}).
		Where("status = ? AND paid_at >= ? AND paid_at < ?", "paid", startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TodayRevenue)

	// Calculate month range in UTC
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	// Month revenue (using UTC range for index efficiency)
	database.DB.Model(&models.PaymentOrder{}).
		Where("status = ? AND paid_at >= ? AND paid_at < ?", "paid", startOfMonth, endOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.MonthRevenue)

	c.JSON(http.StatusOK, stats)
}

// AdminListUserPackages lists all user packages
func AdminListUserPackages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	userID := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.UserPackage{})

	// Filter by status
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by user_id
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get user packages with user info
	var userPackages []models.UserPackage
	err := query.
		Preload("User").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&userPackages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response
	type UserPackageResponse struct {
		ID           string  `json:"id"`
		UserID       string  `json:"user_id"`
		UserEmail    string  `json:"user_email"`
		PackageID    uint    `json:"package_id"`
		PackageName  string  `json:"package_name"`
		PackagePrice float64 `json:"package_price"`
		DurationDays int     `json:"duration_days"`
		DailyLimit   float64 `json:"daily_limit"`
		StartDate    string  `json:"start_date"`
		EndDate      string  `json:"end_date"`
		Status       string  `json:"status"`
		CreatedAt    string  `json:"created_at"`
	}

	var response []UserPackageResponse
	for _, up := range userPackages {
		response = append(response, UserPackageResponse{
			ID:           up.ID.String(),
			UserID:       up.UserID.String(),
			UserEmail:    up.User.Email,
			PackageID:    up.PackageID,
			PackageName:  up.PackageName,
			PackagePrice: up.PackagePrice,
			DurationDays: up.DurationDays,
			DailyLimit:   up.DailyLimit,
			StartDate:    up.StartDate.Format("2006-01-02"),
			EndDate:      up.EndDate.Format("2006-01-02"),
			Status:       up.Status,
			CreatedAt:    up.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_packages": response,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}
