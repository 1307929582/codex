package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
)

var (
	lastUsedUpdate         sync.Map
	lastUsedUpdateInterval = 5 * time.Minute
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		apiKey := parts[1]
		keyHash := HashAPIKey(apiKey)

		var dbKey models.APIKey
		if err := database.DB.Preload("User").Where("key_hash = ? AND status = ?", keyHash, "active").First(&dbKey).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or inactive API key"})
			c.Abort()
			return
		}

		if dbKey.User.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{"error": "user account is not active"})
			c.Abort()
			return
		}

		if dbKey.QuotaLimit != nil && float64(dbKey.TotalUsage) >= *dbKey.QuotaLimit {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "api key quota exceeded"})
			c.Abort()
			return
		}

		if dbKey.User.Balance <= 0 {
			// Check for active package
			today := database.GetToday()
			var activePackage models.UserPackage
			hasActivePackage := database.DB.Where("user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
				dbKey.User.ID, "active", today, today).
				First(&activePackage).Error == nil

			if !hasActivePackage {
				c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance or active package"})
				c.Abort()
				return
			}
		}

		// Update last_used_at asynchronously with throttling
		// Use conditional update to prevent stampede writes
		now := time.Now()
		if shouldUpdateLastUsed(dbKey.ID, now) {
			go func(keyID uint, ts time.Time) {
				// Conditional update: only update if last_used_at is NULL or older than interval
				database.DB.Model(&models.APIKey{}).
					Where("id = ? AND (last_used_at IS NULL OR last_used_at < ?)", keyID, ts.Add(-lastUsedUpdateInterval)).
					Update("last_used_at", ts)
			}(dbKey.ID, now)
		}

		c.Set("user", dbKey.User)
		c.Set("api_key", dbKey)
		c.Next()
	}
}

// HashAPIKey exports the hash function for use in key creation
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func shouldUpdateLastUsed(keyID uint, now time.Time) bool {
	if value, ok := lastUsedUpdate.Load(keyID); ok {
		if last, ok := value.(int64); ok {
			if now.Sub(time.Unix(0, last)) < lastUsedUpdateInterval {
				return false
			}
		}
	}
	lastUsedUpdate.Store(keyID, now.UnixNano())
	return true
}
