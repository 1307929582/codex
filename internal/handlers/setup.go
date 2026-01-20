package handlers

import (
	"net/http"
	"time"

	"codex-gateway/internal/config"
	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SetupStatus checks if initial setup is needed
func SetupStatus(c *gin.Context) {
	var adminCount int64
	database.DB.Model(&models.User{}).
		Where("role IN ?", []string{"admin", "super_admin"}).
		Count(&adminCount)

	c.JSON(http.StatusOK, gin.H{
		"needs_setup": adminCount == 0,
	})
}

// SetupInitialize performs initial setup
func SetupInitialize(c *gin.Context) {
	// Check if already initialized
	var adminCount int64
	database.DB.Model(&models.User{}).
		Where("role IN ?", []string{"admin", "super_admin"}).
		Count(&adminCount)

	if adminCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "system already initialized"})
		return
	}

	var req struct {
		Email                      string  `json:"email" binding:"required,email"`
		Password                   string  `json:"password" binding:"required,min=6"`
		OpenAIAPIKey               string  `json:"openai_api_key" binding:"required"`
		OpenAIBaseURL              string  `json:"openai_base_url"`
		Announcement               string  `json:"announcement"`
		DefaultBalance             float64 `json:"default_balance"`
		EmailRegistrationEnabled   bool    `json:"email_registration_enabled"`
		LinuxDoRegistrationEnabled bool    `json:"linuxdo_registration_enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Create admin user
		admin := models.User{
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			Balance:      0,
			Status:       "active",
			Role:         "super_admin",
		}

		if err := tx.Create(&admin).Error; err != nil {
			return err
		}

		// Create system settings
		settings := models.SystemSettings{
			ID:                         1,
			Announcement:               req.Announcement,
			DefaultBalance:             req.DefaultBalance,
			MinRechargeAmount:          10,
			EmailRegistrationEnabled:   req.EmailRegistrationEnabled,
			LinuxDoRegistrationEnabled: req.LinuxDoRegistrationEnabled,
			OpenAIAPIKey:               req.OpenAIAPIKey,
			OpenAIBaseURL:              req.OpenAIBaseURL,
		}

		if settings.OpenAIBaseURL == "" {
			settings.OpenAIBaseURL = "https://api.openai.com/v1"
		}

		return tx.Create(&settings).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate token for auto-login
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "setup completed successfully",
		"token":   tokenString,
	})
}
