package billing

import (
	"fmt"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeductCost deducts cost from user's package quota or balance
func DeductCost(tx *gorm.DB, userID uuid.UUID, cost float64) error {
	if cost <= 0 {
		return nil
	}

	// Get today's date
	today := database.GetToday()

	// Ensure daily usage record exists for total usage tracking
	dailyUsage := models.DailyUsage{
		UserID:          userID,
		Date:            today,
		UsedAmount:      0,
		TotalUsedAmount: 0,
	}
	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "date"}},
		DoNothing: true,
	}).Create(&dailyUsage).Error
	if err != nil {
		return fmt.Errorf("failed to upsert daily usage: %v", err)
	}

	// Check user daily usage limit
	var user models.User
	if err := tx.Select("daily_usage_limit").Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("failed to get user daily limit: %v", err)
	}

	if user.DailyUsageLimit != nil {
		limit := *user.DailyUsageLimit
		result := tx.Model(&models.DailyUsage{}).
			Where("user_id = ? AND date = ? AND total_used_amount + ? <= ?", userID, today, cost, limit).
			Update("total_used_amount", gorm.Expr("total_used_amount + ?", cost))
		if result.Error != nil {
			return fmt.Errorf("failed to update daily total usage: %v", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("daily usage limit exceeded")
		}
	} else {
		result := tx.Model(&models.DailyUsage{}).
			Where("user_id = ? AND date = ?", userID, today).
			Update("total_used_amount", gorm.Expr("total_used_amount + ?", cost))
		if result.Error != nil {
			return fmt.Errorf("failed to update daily total usage: %v", result.Error)
		}
	}

	// Check if user has active package
	var activePackage models.UserPackage
	err = tx.Where("user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
		userID, "active", today, today).
		Order("end_date ASC").
		First(&activePackage).Error

	if err == nil {
		// User has active package, try to use package quota first
		// Fetch the record (either newly created or existing)
		err = tx.Where("user_id = ? AND date = ?", userID, today).First(&dailyUsage).Error
		if err != nil {
			return fmt.Errorf("failed to get daily usage: %v", err)
		}

		if dailyUsage.UserPackageID == nil || *dailyUsage.UserPackageID != activePackage.ID {
			if err := tx.Model(&models.DailyUsage{}).
				Where("id = ?", dailyUsage.ID).
				Update("user_package_id", activePackage.ID).Error; err != nil {
				return fmt.Errorf("failed to update daily usage package: %v", err)
			}
		}

		// Calculate remaining quota
		remaining := activePackage.DailyLimit - dailyUsage.UsedAmount

		if remaining >= cost {
			// Package quota is enough - use atomic update
			result := tx.Model(&models.DailyUsage{}).
				Where("id = ? AND used_amount + ? <= ?", dailyUsage.ID, cost, activePackage.DailyLimit).
				Update("used_amount", gorm.Expr("used_amount + ?", cost))

			if result.Error != nil {
				return fmt.Errorf("failed to update daily usage: %v", result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("concurrent update conflict or quota exceeded")
			}
			return nil
		} else if remaining > 0 {
			// Use remaining quota, then deduct from balance
			// Use atomic update with condition to prevent concurrent over-use
			result := tx.Model(&models.DailyUsage{}).
				Where("id = ? AND used_amount + ? <= ?", dailyUsage.ID, remaining, activePackage.DailyLimit).
				Update("used_amount", gorm.Expr("used_amount + ?", remaining))

			if result.Error != nil {
				return fmt.Errorf("failed to update daily usage: %v", result.Error)
			}
			if result.RowsAffected == 0 {
				// Quota was consumed by concurrent request, fall back to balance only
				// Don't return error, just use balance for full cost
			} else {
				// Successfully used remaining quota, reduce cost
				cost -= remaining
			}
		}
	}

	// Deduct from balance using atomic operation
	result := tx.Exec("UPDATE users SET balance = balance - ? WHERE id = ? AND balance >= ?", cost, userID, cost)
	if result.Error != nil {
		return fmt.Errorf("failed to update balance: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("insufficient balance")
	}

	return nil
}

// CheckAndExpirePackages checks and expires packages that have passed their end date
func CheckAndExpirePackages() error {
	today := database.GetToday()

	result := database.DB.Model(&models.UserPackage{}).
		Where("status = ? AND end_date < ?", "active", today).
		Update("status", "expired")

	if result.Error != nil {
		return fmt.Errorf("failed to expire packages: %v", result.Error)
	}

	if result.RowsAffected > 0 {
		fmt.Printf("Expired %d packages\n", result.RowsAffected)
	}

	return nil
}

// StartPackageExpirationJob starts a background job to check and expire packages
func StartPackageExpirationJob() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			if err := CheckAndExpirePackages(); err != nil {
				fmt.Printf("Error expiring packages: %v\n", err)
			}
		}
	}()
}
