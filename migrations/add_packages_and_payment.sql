-- Add Credit payment settings to system_settings
ALTER TABLE system_settings
ADD COLUMN IF NOT EXISTS credit_enabled BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS credit_pid VARCHAR(255),
ADD COLUMN IF NOT EXISTS credit_key VARCHAR(255),
ADD COLUMN IF NOT EXISTS credit_notify_url VARCHAR(500),
ADD COLUMN IF NOT EXISTS credit_return_url VARCHAR(500);

-- Create packages table (套餐配置)
CREATE TABLE IF NOT EXISTS packages (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(18,6) NOT NULL,
    duration_days INTEGER NOT NULL,
    daily_limit DECIMAL(18,6) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create user_packages table (用户购买的套餐)
CREATE TABLE IF NOT EXISTS user_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    package_id INTEGER NOT NULL REFERENCES packages(id),
    package_name VARCHAR(100) NOT NULL,
    package_price DECIMAL(18,6) NOT NULL,
    duration_days INTEGER NOT NULL,
    daily_limit DECIMAL(18,6) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_packages_user_status ON user_packages(user_id, status);
CREATE INDEX IF NOT EXISTS idx_user_packages_dates ON user_packages(start_date, end_date);

-- Create daily_usage table (每日使用记录)
CREATE TABLE IF NOT EXISTS daily_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_package_id UUID REFERENCES user_packages(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    used_amount DECIMAL(18,6) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_daily_usage_user_date ON daily_usage(user_id, date);

-- Create payment_orders table (支付订单)
CREATE TABLE IF NOT EXISTS payment_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    package_id INTEGER REFERENCES packages(id) ON DELETE SET NULL,
    order_no VARCHAR(64) UNIQUE NOT NULL,
    out_trade_no VARCHAR(64),
    trade_no VARCHAR(64),
    amount DECIMAL(18,6) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50) DEFAULT 'credit',
    payment_data TEXT,
    notify_data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_orders_user ON payment_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status ON payment_orders(status);
CREATE INDEX IF NOT EXISTS idx_payment_orders_order_no ON payment_orders(order_no);

-- Insert default packages
INSERT INTO packages (name, description, price, duration_days, daily_limit, sort_order) VALUES
('基础套餐', '适合轻度使用，每天$5额度', 30.00, 30, 5.00, 1),
('标准套餐', '适合日常使用，每天$10额度', 50.00, 30, 10.00, 2),
('高级套餐', '适合重度使用，每天$20额度', 90.00, 30, 20.00, 3)
ON CONFLICT DO NOTHING;
