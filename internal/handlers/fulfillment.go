package handlers

import (
	"fmt"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func decrementPackageStock(tx *gorm.DB, pkg *models.Package) error {
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
			return newUserError("套餐库存不足")
		}
		return nil
	}

	return tx.Model(&models.Package{}).
		Where("id = ?", pkg.ID).
		Update("sold_count", gorm.Expr("sold_count + 1")).Error
}

func createUserPackage(tx *gorm.DB, userID uuid.UUID, pkg *models.Package) (*models.UserPackage, error) {
	startDate := time.Now().In(database.AsiaShanghai)
	endDate := startDate.AddDate(0, 0, pkg.DurationDays)

	userPackage := models.UserPackage{
		UserID:       userID,
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
		return nil, err
	}

	return &userPackage, nil
}

func fulfillPackagePurchase(tx *gorm.DB, order *models.PaymentOrder) error {
	if order.PackageID == nil {
		return fmt.Errorf("missing package ID")
	}

	var pkg models.Package
	if err := tx.First(&pkg, order.PackageID).Error; err != nil {
		return err
	}

	if err := decrementPackageStock(tx, &pkg); err != nil {
		return err
	}

	if _, err := createUserPackage(tx, order.UserID, &pkg); err != nil {
		return err
	}

	description := fmt.Sprintf("购买套餐: %s", pkg.Name)
	if order.DiscountAmount > 0 && order.CouponCode != "" {
		description = fmt.Sprintf("%s (优惠码 %s 抵扣 $%.2f)", description, order.CouponCode, order.DiscountAmount)
	}

	transaction := models.Transaction{
		UserID:      order.UserID,
		Amount:      order.Amount,
		Type:        "package_purchase",
		Description: description,
	}

	return tx.Create(&transaction).Error
}

func fulfillPackageSwitch(tx *gorm.DB, order *models.PaymentOrder) error {
	if order.PackageID == nil {
		return fmt.Errorf("missing package ID")
	}
	if order.SwitchFromUserPackageID == nil {
		return fmt.Errorf("missing switch source package")
	}

	var currentPackage models.UserPackage
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ?", order.SwitchFromUserPackageID, order.UserID).
		First(&currentPackage).Error; err != nil {
		return err
	}

	today := database.GetToday()
	if err := tx.Model(&models.UserPackage{}).
		Where("id = ?", currentPackage.ID).
		Updates(map[string]interface{}{
			"status":   "switched",
			"end_date": today,
		}).Error; err != nil {
		return err
	}

	var pkg models.Package
	if err := tx.First(&pkg, order.PackageID).Error; err != nil {
		return err
	}

	if err := decrementPackageStock(tx, &pkg); err != nil {
		return err
	}

	if _, err := createUserPackage(tx, order.UserID, &pkg); err != nil {
		return err
	}

	description := fmt.Sprintf("套餐切换: %s -> %s", currentPackage.PackageName, pkg.Name)
	if order.ProrationCredit > 0 {
		description = fmt.Sprintf("%s (按天折算抵扣 $%.2f)", description, order.ProrationCredit)
	}
	if order.DiscountAmount > 0 && order.CouponCode != "" {
		description = fmt.Sprintf("%s (优惠码 %s 抵扣 $%.2f)", description, order.CouponCode, order.DiscountAmount)
	}

	transaction := models.Transaction{
		UserID:      order.UserID,
		Amount:      order.Amount,
		Type:        "package_switch",
		Description: description,
	}

	return tx.Create(&transaction).Error
}
