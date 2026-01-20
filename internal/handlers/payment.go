package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PurchasePackage creates a payment order for package purchase
func PurchasePackage(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	packageID := c.Param("id")

	var pkg models.Package
	if err := database.DB.First(&pkg, packageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	if pkg.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package is not available"})
		return
	}

	// Check stock availability
	if pkg.Stock != -1 && pkg.Stock <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package is out of stock"})
		return
	}

	// Get system settings for Credit config
	var settings models.SystemSettings
	if err := database.DB.First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	if !settings.CreditEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment is not enabled"})
		return
	}

	// Create payment order
	orderNo := fmt.Sprintf("PKG%d%s", time.Now().Unix(), uuid.New().String()[:8])
	order := models.PaymentOrder{
		UserID:        user.ID,
		PackageID:     &pkg.ID,
		OrderNo:       orderNo,
		Amount:        pkg.Price,
		Status:        "pending",
		PaymentMethod: "credit",
	}

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	// Generate Credit payment URL
	params := map[string]string{
		"pid":           settings.CreditPID,
		"type":          "epay",
		"out_trade_no":  orderNo,
		"name":          pkg.Name,
		"money":         fmt.Sprintf("%.2f", pkg.Price),
		"notify_url":    settings.CreditNotifyURL,
		"return_url":    settings.CreditReturnURL,
	}

	sign := generateCreditSign(params, settings.CreditKey)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	// Build payment URL
	paymentURL := "https://credit.linux.do/epay/pay/submit.php"

	c.JSON(http.StatusOK, gin.H{
		"order_no":    orderNo,
		"amount":      pkg.Price,
		"payment_url": paymentURL,
		"params":      params,
	})
}

// CreditNotify handles Credit payment callback
func CreditNotify(c *gin.Context) {
	// Get all query parameters
	params := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// Get system settings
	var settings models.SystemSettings
	if err := database.DB.First(&settings).Error; err != nil {
		c.String(http.StatusInternalServerError, "error")
		return
	}

	// Verify signature
	sign := params["sign"]
	delete(params, "sign")
	delete(params, "sign_type")

	expectedSign := generateCreditSign(params, settings.CreditKey)
	if sign != expectedSign {
		log.Printf("[Payment] Invalid signature from IP: %s", c.ClientIP())
		c.String(http.StatusBadRequest, "invalid signature")
		return
	}

	// Check trade status
	if params["trade_status"] != "TRADE_SUCCESS" {
		c.String(http.StatusOK, "success")
		return
	}

	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]

	// Validate trade_no is not empty
	if tradeNo == "" {
		log.Printf("[Payment] Empty trade_no from IP: %s", c.ClientIP())
		c.String(http.StatusBadRequest, "invalid trade_no")
		return
	}

	// Find order
	var order models.PaymentOrder
	if err := database.DB.Where("order_no = ?", outTradeNo).First(&order).Error; err != nil {
		log.Printf("[Payment] Order not found: %s from IP: %s", outTradeNo, c.ClientIP())
		c.String(http.StatusNotFound, "order not found")
		return
	}

	// Check if already processed (idempotency)
	if order.Status == "paid" {
		log.Printf("[Payment] Order already paid: %s", outTradeNo)
		c.String(http.StatusOK, "success")
		return
	}

	// Check if order is too old (prevent replay attacks)
	if time.Since(order.CreatedAt) > 24*time.Hour {
		log.Printf("[Payment] Order too old: %s, created at: %s", outTradeNo, order.CreatedAt)
		c.String(http.StatusBadRequest, "order expired")
		return
	}

	// Process payment in transaction
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Update order status
		now := time.Now()
		order.Status = "paid"
		order.TradeNo = tradeNo
		order.PaidAt = &now
		order.NotifyData = c.Request.URL.RawQuery

		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		// Check if this is a recharge order (no package) or package purchase
		if order.PackageID == nil {
			// This is a balance recharge order
			// Add balance to user account
			result := tx.Model(&models.User{}).
				Where("id = ?", order.UserID).
				Update("balance", gorm.Expr("balance + ?", order.Amount))

			if result.Error != nil {
				return fmt.Errorf("failed to update balance: %v", result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("user not found")
			}

			// Create transaction record
			transaction := models.Transaction{
				UserID:      order.UserID,
				Amount:      order.Amount,
				Type:        "deposit",
				Description: fmt.Sprintf("余额充值 $%.2f", order.Amount),
			}

			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}

			log.Printf("[Payment] Balance recharged: user=%s, amount=%.2f, order=%s", order.UserID, order.Amount, order.OrderNo)
		} else {
			// This is a package purchase order
			// Get package info
			var pkg models.Package
			if err := tx.First(&pkg, order.PackageID).Error; err != nil {
				return err
			}

			// Check and decrement stock (atomic operation)
			if pkg.Stock != -1 {
				result := tx.Model(&models.Package{}).
					Where("id = ? AND (stock = -1 OR stock > 0)", pkg.ID).
					Updates(map[string]interface{}{
						"stock":      gorm.Expr("CASE WHEN stock = -1 THEN -1 ELSE stock - 1 END"),
						"sold_count": gorm.Expr("sold_count + 1"),
					})

				if result.Error != nil {
					return fmt.Errorf("failed to update stock: %v", result.Error)
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("package is out of stock")
				}
			} else {
				// Unlimited stock, just increment sold count
				tx.Model(&models.Package{}).Where("id = ?", pkg.ID).
					Update("sold_count", gorm.Expr("sold_count + 1"))
			}

			// Create user package
			startDate := time.Now().In(database.AsiaShanghai)
			endDate := startDate.AddDate(0, 0, pkg.DurationDays)

			userPackage := models.UserPackage{
				UserID:       order.UserID,
				PackageID:    pkg.ID,
				PackageName:  pkg.Name,
				PackagePrice: pkg.Price,
				DurationDays: pkg.DurationDays,
				DailyLimit:   pkg.DailyLimit,
				StartDate:    startDate,
				EndDate:      endDate,
				Status:       "active",
			}

			if err := tx.Create(&userPackage).Error; err != nil {
				return err
			}

			// Create transaction record
			transaction := models.Transaction{
				UserID:      order.UserID,
				Amount:      pkg.Price,
				Type:        "package_purchase",
				Description: fmt.Sprintf("购买套餐: %s", pkg.Name),
			}

			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}

			log.Printf("[Payment] Package purchased: user=%s, package=%s, order=%s", order.UserID, pkg.Name, order.OrderNo)
		}

		return nil
	})

	if err != nil {
		c.String(http.StatusInternalServerError, "error")
		return
	}

	c.String(http.StatusOK, "success")
}

// CreditReturn handles Credit payment return
func CreditReturn(c *gin.Context) {
	orderNo := c.Query("out_trade_no")

	var order models.PaymentOrder
	if err := database.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		// Redirect to account page for recharge orders, packages page for package orders
		c.Redirect(http.StatusFound, "/account?error=order_not_found")
		return
	}

	if order.Status == "paid" {
		// Redirect based on order type
		if order.PackageID == nil {
			// Recharge order - redirect to account page
			c.Redirect(http.StatusFound, "/account?success=recharge_success")
		} else {
			// Package order - redirect to packages page
			c.Redirect(http.StatusFound, "/packages?success=true")
		}
	} else {
		// Redirect based on order type
		if order.PackageID == nil {
			c.Redirect(http.StatusFound, "/account?error=payment_failed")
		} else {
			c.Redirect(http.StatusFound, "/packages?error=payment_failed")
		}
	}
}

// CreateRechargeOrder creates a payment order for balance recharge
func CreateRechargeOrder(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get system settings for Credit config and min recharge amount
	var settings models.SystemSettings
	if err := database.DB.First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	if !settings.CreditEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment is not enabled"})
		return
	}

	// Check minimum recharge amount
	if req.Amount < settings.MinRechargeAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("minimum recharge amount is $%.2f", settings.MinRechargeAmount),
		})
		return
	}

	// Create payment order
	orderNo := fmt.Sprintf("RCH%d%s", time.Now().Unix(), uuid.New().String()[:8])
	order := models.PaymentOrder{
		UserID:        user.ID,
		PackageID:     nil, // No package for recharge
		OrderNo:       orderNo,
		Amount:        req.Amount,
		Status:        "pending",
		PaymentMethod: "credit",
	}

	if err := database.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	// Generate Credit payment URL
	params := map[string]string{
		"pid":          settings.CreditPID,
		"type":         "epay",
		"out_trade_no": orderNo,
		"name":         fmt.Sprintf("余额充值 $%.2f", req.Amount),
		"money":        fmt.Sprintf("%.2f", req.Amount),
		"notify_url":   settings.CreditNotifyURL,
		"return_url":   settings.CreditReturnURL,
	}

	sign := generateCreditSign(params, settings.CreditKey)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	// Build payment URL
	paymentURL := "https://credit.linux.do/epay/pay/submit.php"

	c.JSON(http.StatusOK, gin.H{
		"order_no":    orderNo,
		"amount":      req.Amount,
		"payment_url": paymentURL,
		"params":      params,
	})
}

// generateCreditSign generates MD5 signature for Credit payment
func generateCreditSign(params map[string]string, key string) string {
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "sign_type" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build string
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	signStr := strings.Join(parts, "&") + key

	// MD5 hash
	hash := md5.Sum([]byte(signStr))
	return hex.EncodeToString(hash[:])
}
