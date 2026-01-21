package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminListCoupons lists coupons with pagination
func AdminListCoupons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Coupon{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("code ILIKE ?", like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var coupons []models.Coupon
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&coupons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch coupons"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coupons": coupons,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// AdminCreateCoupon creates a new coupon
func AdminCreateCoupon(c *gin.Context) {
	var req struct {
		Code      string  `json:"code" binding:"required"`
		Type      string  `json:"type" binding:"required,oneof=fixed percent"`
		Value     float64 `json:"value" binding:"required,gt=0"`
		MaxUses   int     `json:"max_uses"`
		MinAmount float64 `json:"min_amount"`
		StartsAt  string  `json:"starts_at"`
		EndsAt    string  `json:"ends_at"`
		Status    string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	code := normalizeCouponCode(req.Code)
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "coupon code is required"})
		return
	}

	if req.Type == "percent" && req.Value > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "percent value must be <= 100"})
		return
	}

	startsAt, err := parseCouponTime(req.StartsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	endsAt, err := parseCouponTime(req.EndsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if startsAt != nil && endsAt != nil && endsAt.Before(*startsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ends_at must be after starts_at"})
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "inactive" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	coupon := models.Coupon{
		Code:      code,
		Type:      req.Type,
		Value:     req.Value,
		MaxUses:   req.MaxUses,
		MinAmount: req.MinAmount,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		Status:    status,
	}

	if err := database.DB.Create(&coupon).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create coupon"})
		return
	}

	c.JSON(http.StatusOK, coupon)
}

// AdminUpdateCoupon updates an existing coupon
func AdminUpdateCoupon(c *gin.Context) {
	id := c.Param("id")

	var coupon models.Coupon
	if err := database.DB.First(&coupon, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "coupon not found"})
		return
	}

	var req struct {
		Code      *string  `json:"code"`
		Type      *string  `json:"type"`
		Value     *float64 `json:"value"`
		MaxUses   *int     `json:"max_uses"`
		MinAmount *float64 `json:"min_amount"`
		StartsAt  *string  `json:"starts_at"`
		EndsAt    *string  `json:"ends_at"`
		Status    *string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Code != nil {
		normalized := normalizeCouponCode(*req.Code)
		if normalized == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "coupon code is required"})
			return
		}
		coupon.Code = normalized
	}
	if req.Type != nil {
		if *req.Type != "fixed" && *req.Type != "percent" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon type"})
			return
		}
		coupon.Type = *req.Type
	}
	if req.Value != nil {
		if *req.Value <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon value"})
			return
		}
		if coupon.Type == "percent" && *req.Value > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "percent value must be <= 100"})
			return
		}
		coupon.Value = *req.Value
	}
	if req.MaxUses != nil {
		coupon.MaxUses = *req.MaxUses
	}
	if req.MinAmount != nil {
		coupon.MinAmount = *req.MinAmount
	}
	if req.StartsAt != nil {
		if *req.StartsAt == "" {
			coupon.StartsAt = nil
		} else {
			parsed, err := parseCouponTime(*req.StartsAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			coupon.StartsAt = parsed
		}
	}
	if req.EndsAt != nil {
		if *req.EndsAt == "" {
			coupon.EndsAt = nil
		} else {
			parsed, err := parseCouponTime(*req.EndsAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			coupon.EndsAt = parsed
		}
	}
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "inactive" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		coupon.Status = *req.Status
	}

	if coupon.Type == "percent" && coupon.Value > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "percent value must be <= 100"})
		return
	}

	if coupon.StartsAt != nil && coupon.EndsAt != nil && coupon.EndsAt.Before(*coupon.StartsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ends_at must be after starts_at"})
		return
	}

	if err := database.DB.Save(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update coupon"})
		return
	}

	c.JSON(http.StatusOK, coupon)
}
