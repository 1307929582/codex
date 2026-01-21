package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codex-gateway/internal/billing"
	"codex-gateway/internal/config"
	"codex-gateway/internal/database"
	"codex-gateway/internal/handlers"
	"codex-gateway/internal/middleware"
	"codex-gateway/internal/pricing"
	"codex-gateway/internal/upstream"

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

	// Run custom migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	if err := database.SeedDefaultPricing(); err != nil {
		log.Fatal("Failed to seed default pricing:", err)
	}

	if err := database.SeedCodexPricing(); err != nil {
		log.Fatal("Failed to seed Codex pricing:", err)
	}

	if err := database.SeedCodexUpstreams(); err != nil {
		log.Fatal("Failed to seed Codex upstreams:", err)
	}

	// Initialize pricing service
	pricingService := pricing.GetService()
	if err := pricingService.Initialize(); err != nil {
		log.Printf("Warning: Pricing service failed to initialize: %v", err)
	}
	defer pricingService.Stop()

	// Initialize upstream health checker
	healthChecker := upstream.GetHealthChecker()
	healthChecker.Start()
	defer healthChecker.Stop()

	// Start package expiration job
	billing.StartPackageExpirationJob()
	log.Println("Package expiration job started")

	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Setup Routes (Public - for initial setup only)
	setupGroup := router.Group("/api/setup")
	{
		setupGroup.GET("/status", handlers.SetupStatus)
		setupGroup.POST("/initialize", handlers.SetupInitialize)
	}

	// Control Plane API
	apiGroup := router.Group("/api")
	{
		// Auth Routes (Public)
		auth := apiGroup.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)

			// LinuxDo OAuth
			auth.GET("/linuxdo", handlers.LinuxDoLogin)
			auth.GET("/linuxdo/callback", handlers.LinuxDoCallback)
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
			protected.GET("/usage/daily-trend", handlers.GetDailyTrend)
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
			admin.GET("/stats/usage-chart", handlers.AdminGetUsageChart)

			// Logs
			admin.GET("/logs", handlers.AdminGetLogs)
			admin.GET("/usage/logs", handlers.AdminGetUsageLogs)

			// Pricing Service Status
			admin.GET("/pricing/status", handlers.AdminGetPricingStatus)
			admin.POST("/pricing/reset", handlers.AdminResetPricing)
			admin.GET("/pricing", handlers.AdminListPricing)
			admin.PUT("/pricing/:id", handlers.AdminUpdatePricing)
			admin.POST("/pricing/batch-update-markup", handlers.AdminBatchUpdateMarkup)

			// Package Management
			admin.GET("/packages", handlers.AdminListPackages)
			admin.POST("/packages", handlers.AdminCreatePackage)
			admin.PUT("/packages/:id", handlers.AdminUpdatePackage)
			admin.DELETE("/packages/:id", handlers.AdminDeletePackage)
			admin.PUT("/packages/:id/status", handlers.AdminUpdatePackageStatus)

			// Order Management
			admin.GET("/orders", handlers.AdminListOrders)
			admin.GET("/orders/stats", handlers.AdminGetOrderStats)
			admin.GET("/user-packages", handlers.AdminListUserPackages)

			// Codex Upstream Management
			admin.GET("/codex/upstreams", handlers.AdminListCodexUpstreams)
			admin.GET("/codex/upstreams/:id", handlers.AdminGetCodexUpstream)
			admin.POST("/codex/upstreams", handlers.AdminCreateCodexUpstream)
			admin.PUT("/codex/upstreams/:id", handlers.AdminUpdateCodexUpstream)
			admin.DELETE("/codex/upstreams/:id", handlers.AdminDeleteCodexUpstream)
			admin.PUT("/codex/upstreams/:id/status", handlers.AdminUpdateCodexUpstreamStatus)

			// Upstream Health Check
			admin.GET("/codex/upstreams/health", handlers.AdminGetUpstreamHealth)
			admin.POST("/codex/upstreams/health/check", handlers.AdminTriggerHealthCheck)
		}

		// User Routes (authenticated)
		user := apiGroup.Group("")
		user.Use(middleware.JWTAuthMiddleware())
		{
			// Package Routes
			user.GET("/packages", handlers.ListPackages)
			user.POST("/packages/:id/purchase", handlers.PurchasePackage)
			user.GET("/user/packages", handlers.GetUserPackages)
			user.GET("/user/daily-usage", handlers.GetUserDailyUsage)

			// Recharge Routes
			user.POST("/recharge", handlers.CreateRechargeOrder)
		}

		// Payment Callback Routes (no auth required)
		apiGroup.GET("/payment/credit/notify", handlers.CreditNotify)
		apiGroup.GET("/payment/credit/notify ", handlers.CreditNotify) // Handle URL with trailing space
		apiGroup.GET("/payment/credit/return", handlers.CreditReturn)
	}

	// Data Plane API (OpenAI/Codex Proxy)
	api := router.Group("/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// ChatGPT API
		api.POST("/chat/completions", handlers.ProxyHandler)

		// Codex API (GitHub Copilot)
		api.POST("/completions", handlers.ProxyHandler)
		api.POST("/responses", handlers.ProxyHandler)
		api.POST("/engines/:engine/completions", handlers.ProxyHandler)

		// Other OpenAI APIs
		api.POST("/edits", handlers.ProxyHandler)
		api.POST("/embeddings", handlers.ProxyHandler)
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
