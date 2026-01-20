package billing

import (
	"fmt"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeductCost deducts cost from user's package quota or balance
func DeductCost(tx *gorm.DB, userID uuid.UUID, cost float64) error {
	if cost <= 0 {
		return nil
	}

	// Get today's date
	today := database.GetToday()

	// Check if user has active package
	var activePackage models.UserPackage
	err := tx.Where("user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
		userID, "active", today, today).
		Order("end_date ASC").
		First(&activePackage).Error

	if err == nil {
		// User has active package, try to use package quota first
		var dailyUsage models.DailyUsage
		err = tx.Where("user_id = ? AND date = ?", userID, today).First(&dailyUsage).Error

		if err == gorm.ErrRecordNotFound {
			// Create new daily usage record
			dailyUsage = models.DailyUsage{
				UserID:        userID,
				UserPackageID: &activePackage.ID,
				Date:          today,
				UsedAmount:    0,
			}
			if err := tx.Create(&dailyUsage).Error; err != nil {
				return fmt.Errorf("failed to create daily usage: %v", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to get daily usage: %v", err)
		}

		// Calculate remaining quota
		remaining := activePackage.DailyLimit - dailyUsage.UsedAmount

		if remaining >= cost {
			// Package quota is enough
			dailyUsage.UsedAmount += cost
			if err := tx.Save(&dailyUsage).Error; err != nil {
				return fmt.Errorf("failed to update daily usage: %v", err)
			}
			return nil
		} else if remaining > 0 {
			// Use remaining quota, then deduct from balance
			dailyUsage.UsedAmount = activePackage.DailyLimit
			if err := tx.Save(&dailyUsage).Error; err != nil {
				return fmt.Errorf("failed to update daily usage: %v", err)
			}
			cost -= remaining
		}
	}

	// Deduct from balance
	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	if user.Balance < cost {
		return fmt.Errorf("insufficient balance")
	}

	user.Balance -= cost
	if err := tx.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update balance: %v", err)
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
