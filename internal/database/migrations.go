package database

import (
	"log"

	"gorm.io/gorm"
)

// RunMigrations runs all custom database migrations
func RunMigrations() error {
	log.Println("Running database migrations...")

	// Migration 002: Add OAuth fields to users table
	if err := migration002AddOAuthFields(); err != nil {
		return err
	}

	// Migration 003: Add LinuxDo OAuth settings
	if err := migration003AddLinuxDoOAuthSettings(); err != nil {
		return err
	}

	// Migration 004: Add performance indexes
	if err := migration004AddPerformanceIndexes(); err != nil {
		return err
	}

	log.Println("All migrations completed successfully")
	return nil
}

// migration002AddOAuthFields adds OAuth fields to users table
func migration002AddOAuthFields() error {
	log.Println("Running migration 002: Add OAuth fields to users table")

	// Check if oauth_provider column exists
	var count int64
	DB.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'oauth_provider'").Scan(&count)
	if count > 0 {
		log.Println("Migration 002: Already applied, skipping")
		return nil
	}

	// Add OAuth fields
	sqls := []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500)",
		"ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_oauth ON users(oauth_provider, oauth_id)",
		"UPDATE users SET oauth_provider = 'email' WHERE oauth_provider IS NULL",
	}

	for _, sql := range sqls {
		if err := DB.Exec(sql).Error; err != nil {
			log.Printf("Migration 002 failed at: %s, error: %v", sql, err)
			return err
		}
	}

	log.Println("Migration 002: Completed successfully")
	return nil
}

// migration003AddLinuxDoOAuthSettings adds LinuxDo OAuth settings to system_settings
func migration003AddLinuxDoOAuthSettings() error {
	log.Println("Running migration 003: Add LinuxDo OAuth settings")

	// Check if linuxdo_client_id column exists
	var count int64
	DB.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'system_settings' AND column_name = 'linuxdo_client_id'").Scan(&count)
	if count > 0 {
		log.Println("Migration 003: Already applied, skipping")
		return nil
	}

	// Add LinuxDo OAuth fields
	sqls := []string{
		"ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_id VARCHAR(255)",
		"ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_secret VARCHAR(255)",
		"ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_enabled BOOLEAN DEFAULT false",
	}

	for _, sql := range sqls {
		if err := DB.Exec(sql).Error; err != nil {
			log.Printf("Migration 003 failed at: %s, error: %v", sql, err)
			return err
		}
	}

	// Auto-configure LinuxDo OAuth with default credentials
	result := DB.Exec(`
		UPDATE system_settings
		SET
			linuxdo_client_id = 'kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ',
			linuxdo_client_secret = 'XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg',
			linuxdo_enabled = true
		WHERE id = 1 AND (linuxdo_client_id IS NULL OR linuxdo_client_id = '')
	`)

	if result.Error != nil {
		log.Printf("Migration 003: Failed to set default LinuxDo config: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.Println("Migration 003: LinuxDo OAuth auto-configured with default credentials")
	}

	log.Println("Migration 003: Completed successfully")
	return nil
}

// migration004AddPerformanceIndexes adds indexes for high-concurrency queries
func migration004AddPerformanceIndexes() error {
	log.Println("Running migration 004: Add performance indexes")

	db := DB.Session(&gorm.Session{SkipDefaultTransaction: true})

	sqls := []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_keys_key_hash_status ON api_keys(key_hash, status)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_packages_active_user_end_date ON user_packages(user_id, end_date) WHERE status = 'active'",
		"CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_daily_usage_user_date_unique ON daily_usage(user_id, date)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_orders_order_no ON payment_orders(order_no)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_user_created ON usage_logs(user_id, created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_created_at ON usage_logs(created_at DESC)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_orders_paid_at ON payment_orders(paid_at) WHERE status = 'paid'",
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Migration 004 failed at: %s, error: %v", sql, err)
			return err
		}
	}

	log.Println("Migration 004: Completed successfully")
	return nil
}
