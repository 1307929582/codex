-- Performance optimization indexes for high concurrency support
-- These indexes significantly improve query performance for authentication and usage tracking
-- IMPORTANT: These indexes are created with CONCURRENTLY to avoid blocking writes in production

-- Index for API key authentication (most frequent query)
-- Covers: WHERE key_hash = ? AND status = ?
-- Note: key_hash already has a unique index, so this composite index may be redundant
-- Consider removing if profiling shows no benefit
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_keys_key_hash_status
ON api_keys(key_hash, status);

-- Index for user package lookup (used in every API request with active package)
-- Covers: WHERE user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?
-- Optimized with partial index for active packages only
DROP INDEX CONCURRENTLY IF EXISTS idx_user_packages_user_status_dates;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_packages_active_user_end_date
ON user_packages(user_id, end_date)
WHERE status = 'active';

-- Index for daily usage tracking (used in billing operations)
-- Covers: WHERE user_id = ? AND date = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_daily_usage_user_date
ON daily_usage(user_id, date);

-- Index for payment order lookup (used in payment callbacks)
-- Covers: WHERE order_no = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_orders_order_no
ON payment_orders(order_no);

-- Index for usage log queries (used in admin analytics)
-- Covers: WHERE user_id = ? ORDER BY created_at DESC
-- Note: This may be redundant with existing GORM composite index
-- Consider removing if profiling shows no benefit
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_user_created
ON usage_logs(user_id, created_at DESC);

-- Index for payment orders date queries (used in revenue statistics)
-- Covers: WHERE status = 'paid' AND DATE(paid_at) = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_payment_orders_paid_at
ON payment_orders(paid_at)
WHERE status = 'paid';

-- Analyze tables to update statistics for query planner
ANALYZE api_keys;
ANALYZE user_packages;
ANALYZE daily_usage;
ANALYZE payment_orders;
ANALYZE usage_logs;
