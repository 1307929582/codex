package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"codex-gateway/internal/config"
	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// getLinuxDoOAuthConfig returns OAuth config from database settings
func getLinuxDoOAuthConfig() (*oauth2.Config, error) {
	var settings models.SystemSettings
	if err := database.DB.First(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to load system settings: %v", err)
	}

	if !settings.LinuxDoEnabled {
		return nil, fmt.Errorf("LinuxDo OAuth is not enabled")
	}

	if settings.LinuxDoClientID == "" || settings.LinuxDoClientSecret == "" {
		return nil, fmt.Errorf("LinuxDo OAuth credentials not configured")
	}

	return &oauth2.Config{
		ClientID:     settings.LinuxDoClientID,
		ClientSecret: settings.LinuxDoClientSecret,
		RedirectURL:  config.AppConfig.FrontendURL + "/api/auth/linuxdo/callback",
		Scopes:       []string{"read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://linux.do/oauth2/authorize",
			TokenURL: "https://linux.do/oauth2/token",
		},
	}, nil
}

// LinuxDoUserInfo represents user data from LinuxDo API
type LinuxDoUserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar_template"`
}

// LinuxDoLogin initiates OAuth flow
func LinuxDoLogin(c *gin.Context) {
	oauthConfig, err := getLinuxDoOAuthConfig()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	state := generateRandomState()
	// Store state in session/cookie for CSRF protection
	c.SetCookie("oauth_state", state, 600, "/", "", true, true)

	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// LinuxDoCallback handles OAuth callback
func LinuxDoCallback(c *gin.Context) {
	oauthConfig, err := getLinuxDoOAuthConfig()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	// Verify state for CSRF protection
	state := c.Query("state")
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || state != cookieState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}

	// Exchange code for token
	code := c.Query("code")
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to exchange token"})
		return
	}

	// Get user info from LinuxDo
	userInfo, err := getLinuxDoUserInfo(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	// Find or create user
	user, err := findOrCreateLinuxDoUser(userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Generate JWT token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := jwtToken.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Redirect to frontend with token
	frontendURL := config.AppConfig.FrontendURL
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, tokenString))
}

func getLinuxDoUserInfo(accessToken string) (*LinuxDoUserInfo, error) {
	req, err := http.NewRequest("GET", "https://linux.do/api/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Api-Key", accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo LinuxDoUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func findOrCreateLinuxDoUser(userInfo *LinuxDoUserInfo) (*models.User, error) {
	var user models.User

	// Get default balance from settings
	var settings models.SystemSettings
	database.DB.First(&settings)
	defaultBalance := settings.DefaultBalance

	// Try to find existing user by OAuth ID
	err := database.DB.Where("oauth_provider = ? AND oauth_id = ?", "linuxdo", fmt.Sprintf("%d", userInfo.ID)).First(&user).Error
	if err == nil {
		// User exists, update info
		user.Username = userInfo.Username
		user.AvatarURL = fmt.Sprintf("https://linux.do%s", userInfo.Avatar)
		if userInfo.Email != "" {
			user.Email = userInfo.Email
		}
		database.DB.Save(&user)
		return &user, nil
	}

	// Try to find by email (for account linking)
	if userInfo.Email != "" {
		err = database.DB.Where("email = ?", userInfo.Email).First(&user).Error
		if err == nil {
			// Link existing account to LinuxDo
			user.OAuthProvider = "linuxdo"
			user.OAuthID = fmt.Sprintf("%d", userInfo.ID)
			user.Username = userInfo.Username
			user.AvatarURL = fmt.Sprintf("https://linux.do%s", userInfo.Avatar)
			database.DB.Save(&user)
			return &user, nil
		}
	}

	// Create new user
	email := userInfo.Email
	if email == "" {
		// Generate a placeholder email if LinuxDo doesn't provide one
		email = fmt.Sprintf("linuxdo_%d@oauth.local", userInfo.ID)
	}

	user = models.User{
		Email:         email,
		Username:      userInfo.Username,
		OAuthProvider: "linuxdo",
		OAuthID:       fmt.Sprintf("%d", userInfo.ID),
		AvatarURL:     fmt.Sprintf("https://linux.do%s", userInfo.Avatar),
		Balance:       defaultBalance,
		Status:        "active",
		Role:          "user",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func generateRandomState() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
