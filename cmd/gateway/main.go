package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codex-gateway/internal/config"
	"codex-gateway/internal/database"
	"codex-gateway/internal/handlers"
	"codex-gateway/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if err := database.SeedDefaultPricing(); err != nil {
		log.Fatal("Failed to seed default pricing:", err)
	}

	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Control Plane API
	apiGroup := router.Group("/api")
	{
		// Auth Routes (Public)
		auth := apiGroup.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// Protected Routes
		protected := apiGroup.Group("")
		protected.Use(middleware.JWTAuthMiddleware())
		{
			protected.GET("/auth/me", handlers.GetMe)

			// API Keys
			protected.GET("/keys", handlers.ListAPIKeys)
			protected.POST("/keys", handlers.CreateAPIKey)
			protected.DELETE("/keys/:id", handlers.DeleteAPIKey)
			protected.PUT("/keys/:id/status", handlers.UpdateAPIKeyStatus)

			// Usage & Billing
			protected.GET("/usage/logs", handlers.GetUsageLogs)
			protected.GET("/usage/stats", handlers.GetUsageStats)
			protected.GET("/account/balance", handlers.GetBalance)
			protected.GET("/account/transactions", handlers.GetTransactions)
		}

		// Admin Routes
		admin := apiGroup.Group("/admin")
		admin.Use(middleware.JWTAuthMiddleware())
		admin.Use(middleware.AdminAuthMiddleware())
		{
			// User Management
			admin.GET("/users", handlers.AdminListUsers)
			admin.GET("/users/:id", handlers.AdminGetUser)
			admin.PUT("/users/:id/balance", handlers.AdminUpdateBalance)
			admin.PUT("/users/:id/status", handlers.AdminUpdateUserStatus)

			// System Settings
			admin.GET("/settings", handlers.AdminGetSettings)
			admin.PUT("/settings", handlers.AdminUpdateSettings)

			// Statistics
			admin.GET("/stats/overview", handlers.AdminGetOverview)

			// Logs
			admin.GET("/logs", handlers.AdminGetLogs)
		}
	}

	// Data Plane API (OpenAI Proxy)
	api := router.Group("/v1")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/chat/completions", handlers.ProxyHandler)
	}

	log.Printf("Server starting on port %s", config.AppConfig.ServerPort)

	srv := &http.Server{
		Addr:    ":" + config.AppConfig.ServerPort,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
