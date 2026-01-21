package handlers

import (
	"fmt"
	"strings"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func normalizeCouponCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func parseCouponTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04",
	}

	for _, layout := range layouts {
		var parsed time.Time
		var err error
		if layout == time.RFC3339 {
			parsed, err = time.Parse(layout, value)
		} else {
			parsed, err = time.ParseInLocation(layout, value, database.AsiaShanghai)
		}
		if err == nil {
			return &parsed, nil
		}
	}

	return nil, fmt.Errorf("invalid time format")
}

func getCouponForUpdate(tx *gorm.DB, code string) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("code = ?", code).
		First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

func validateCoupon(coupon *models.Coupon, amount float64) error {
	if coupon.Status != "active" {
		return newUserError("优惠码不可用")
	}
	if coupon.Type != "fixed" && coupon.Type != "percent" {
		return newUserError("优惠码配置错误")
	}
	if coupon.Value <= 0 {
		return newUserError("优惠码配置错误")
	}
	if coupon.Type == "percent" && coupon.Value > 100 {
		return newUserError("优惠码折扣比例无效")
	}
	if coupon.MinAmount > 0 && amount < coupon.MinAmount {
		return newUserError(fmt.Sprintf("订单金额需满 $%.2f 才可使用该优惠码", coupon.MinAmount))
	}
	if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
		return newUserError("优惠码已被使用完")
	}
	now := time.Now()
	if coupon.StartsAt != nil && now.Before(*coupon.StartsAt) {
		return newUserError("优惠码尚未生效")
	}
	if coupon.EndsAt != nil && now.After(*coupon.EndsAt) {
		return newUserError("优惠码已过期")
	}
	return nil
}

func computeCouponDiscount(coupon *models.Coupon, amount float64) float64 {
	if coupon == nil || amount <= 0 {
		return 0
	}
	if coupon.Type == "percent" {
		return amount * (coupon.Value / 100.0)
	}
	return coupon.Value
}

func consumeCoupon(tx *gorm.DB, coupon *models.Coupon, userID uuid.UUID, order *models.PaymentOrder, discount float64) error {
	if coupon == nil {
		return nil
	}

	update := tx.Model(&models.Coupon{}).Where("id = ?", coupon.ID)
	if coupon.MaxUses > 0 {
		update = update.Where("used_count < ?", coupon.MaxUses)
	}
	result := update.Update("used_count", gorm.Expr("used_count + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return newUserError("优惠码已被使用完")
	}

	redemption := models.CouponRedemption{
		CouponID:       coupon.ID,
		UserID:         userID,
		OrderID:        &order.ID,
		OrderNo:        order.OrderNo,
		DiscountAmount: discount,
	}
	return tx.Create(&redemption).Error
}
