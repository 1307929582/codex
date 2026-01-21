package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SwitchPackage upgrades/downgrades the current package with daily proration
func SwitchPackage(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	packageID := c.Param("id")

	var req struct {
		CouponCode string `json:"coupon_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var targetPackage models.Package
	if err := database.DB.First(&targetPackage, packageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}
	if targetPackage.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package is not available"})
		return
	}
	if targetPackage.Stock != -1 && targetPackage.Stock <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package is out of stock"})
		return
	}

	var settings models.SystemSettings
	if err := database.DB.First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	var order models.PaymentOrder
	var balanceCredit float64
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		today := database.GetToday()
		var currentPackage models.UserPackage
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
				user.ID, "active", today, today).
			Order("end_date ASC").
			First(&currentPackage).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newUserError("当前没有可切换的套餐")
			}
			return err
		}

		if currentPackage.PackageID == targetPackage.ID {
			return newUserError("已在当前套餐，无需切换")
		}

		remainingDays := calculateRemainingDays(currentPackage.EndDate, today)
		if remainingDays <= 0 {
			return newUserError("当前套餐已过期")
		}

		oldDailyRate := 0.0
		if currentPackage.DurationDays > 0 {
			oldDailyRate = currentPackage.PackagePrice / float64(currentPackage.DurationDays)
		}
		credit := oldDailyRate * float64(remainingDays)
		if credit > currentPackage.PackagePrice {
			credit = currentPackage.PackagePrice
		}
		appliedCredit := credit
		if appliedCredit > targetPackage.Price {
			appliedCredit = targetPackage.Price
		}
		balanceCredit = roundAmount(credit - appliedCredit)
		if balanceCredit < 0 {
			balanceCredit = 0
		}
		payable := roundAmount(targetPackage.Price - appliedCredit)

		couponCode := normalizeCouponCode(req.CouponCode)
		var coupon *models.Coupon
		var discount float64

		if payable > 0 && couponCode != "" {
			var err error
			coupon, err = getCouponForUpdate(tx, couponCode)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return newUserError("优惠码无效")
				}
				return err
			}
			if err := validateCoupon(coupon, payable); err != nil {
				return err
			}
			discount = computeCouponDiscount(coupon, payable)
		}

		payable = roundAmount(payable - discount)
		if payable < 0 {
			payable = 0
		}
		discount = roundAmount(targetPackage.Price - appliedCredit - payable)
		if discount < 0 {
			discount = 0
		}

		if payable > 0 && !settings.CreditEnabled {
			return newUserError("payment is not enabled")
		}

		orderNo := fmt.Sprintf("SWP%d%s", time.Now().Unix(), uuid.New().String()[:8])
		order = models.PaymentOrder{
			UserID:                  user.ID,
			PackageID:               &targetPackage.ID,
			OrderNo:                 orderNo,
			Amount:                  payable,
			OriginalAmount:          roundAmount(targetPackage.Price),
			DiscountAmount:          discount,
			ProrationCredit:         roundAmount(appliedCredit),
			Status:                  "pending",
			PaymentMethod:           "credit",
			OrderType:               "package_switch",
			SwitchFromUserPackageID: &currentPackage.ID,
		}

		if coupon != nil {
			order.CouponID = &coupon.ID
			order.CouponCode = coupon.Code
		}

		if payable <= 0 {
			now := time.Now()
			order.Status = "paid"
			if coupon != nil {
				order.PaymentMethod = "coupon"
			} else {
				order.PaymentMethod = "proration"
			}
			order.PaidAt = &now
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		if coupon != nil {
			if err := consumeCoupon(tx, coupon, user.ID, &order, discount); err != nil {
				return err
			}
		}

		if payable <= 0 {
			if err := fulfillPackageSwitch(tx, &order); err != nil {
				return err
			}

			if balanceCredit > 0 {
				result := tx.Model(&models.User{}).
					Where("id = ?", user.ID).
					Update("balance", gorm.Expr("balance + ?", balanceCredit))
				if result.Error != nil {
					return result.Error
				}

				transaction := models.Transaction{
					UserID:      user.ID,
					Amount:      balanceCredit,
					Type:        "package_switch_credit",
					Description: fmt.Sprintf("套餐折算余额补偿 $%.2f", balanceCredit),
				}
				if err := tx.Create(&transaction).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		if isUserError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create switch order"})
		return
	}

	if order.Amount <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"order_no":       order.OrderNo,
			"amount":         order.Amount,
			"status":         order.Status,
			"balance_credit": balanceCredit,
		})
		return
	}

	// Generate Credit payment URL
	params := map[string]string{
		"pid":          settings.CreditPID,
		"type":         "epay",
		"out_trade_no": order.OrderNo,
		"name":         fmt.Sprintf("套餐切换 %s", targetPackage.Name),
		"money":        fmt.Sprintf("%.2f", order.Amount),
		"notify_url":   settings.CreditNotifyURL,
		"return_url":   settings.CreditReturnURL,
	}

	sign := generateCreditSign(params, settings.CreditKey)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	paymentURL := "https://credit.linux.do/epay/pay/submit.php"

	c.JSON(http.StatusOK, gin.H{
		"order_no":    order.OrderNo,
		"amount":      order.Amount,
		"payment_url": paymentURL,
		"params":      params,
	})
}
