-- Performance optimization indexes for high concurrency support
-- These indexes significantly improve query performance for authentication and usage tracking

-- Index for API key authentication (most frequent query)
-- Covers: WHERE key_hash = ? AND status = ?
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash_status
ON api_keys(key_hash, status);

-- Index for user package lookup (used in every API request with active package)
-- Covers: WHERE user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?
CREATE INDEX IF NOT EXISTS idx_user_packages_user_status_dates
ON user_packages(user_id, status, start_date, end_date);

-- Index for daily usage tracking (used in billing operations)
-- Covers: WHERE user_id = ? AND date = ?
CREATE INDEX IF NOT EXISTS idx_daily_usage_user_date
ON daily_usage(user_id, date);

-- Index for payment order lookup (used in payment callbacks)
-- Covers: WHERE order_no = ?
CREATE INDEX IF NOT EXISTS idx_payment_orders_order_no
ON payment_orders(order_no);

-- Index for usage log queries (used in admin analytics)
-- Covers: WHERE user_id = ? ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_usage_logs_user_created
ON usage_logs(user_id, created_at DESC);

-- Analyze tables to update statistics for query planner
ANALYZE api_keys;
ANALYZE user_packages;
ANALYZE daily_usage;
ANALYZE payment_orders;
ANALYZE usage_logs;
