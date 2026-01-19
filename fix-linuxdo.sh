#!/bin/bash
set -e

echo "=== LinuxDo OAuth 修复脚本 ==="
echo ""

# 1. 添加数据库字段
echo "1. 添加数据库字段..."
docker exec codex-postgres psql -U codex_user -d codex_gateway << 'SQL'
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_id VARCHAR(255);
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_secret VARCHAR(255);
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_enabled BOOLEAN DEFAULT false;

ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
CREATE INDEX IF NOT EXISTS idx_oauth ON users(oauth_provider, oauth_id);
UPDATE users SET oauth_provider = 'email' WHERE oauth_provider IS NULL;
SQL
echo "✓ 字段添加完成"
echo ""

# 2. 配置LinuxDo OAuth
echo "2. 配置LinuxDo OAuth..."
docker exec codex-postgres psql -U codex_user -d codex_gateway << 'SQL'
INSERT INTO system_settings (id, announcement, default_balance, min_recharge_amount, registration_enabled, linuxdo_client_id, linuxdo_client_secret, linuxdo_enabled)
VALUES (1, '', 0, 10, true, 'kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ', 'XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg', true)
ON CONFLICT (id) DO UPDATE SET
  linuxdo_client_id = 'kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ',
  linuxdo_client_secret = 'XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg',
  linuxdo_enabled = true;
SQL
echo "✓ 配置完成"
echo ""

# 3. 重启后端服务
echo "3. 重启后端服务..."
docker restart codex-backend
echo "等待服务启动..."
sleep 8
echo "✓ 服务已重启"
echo ""

# 4. 验证配置
echo "4. 验证配置..."
docker exec codex-postgres psql -U codex_user -d codex_gateway -c "SELECT id, linuxdo_client_id, linuxdo_enabled FROM system_settings;"
echo ""

# 5. 测试OAuth端点
echo "5. 测试LinuxDo OAuth端点..."
curl -s http://localhost:12322/api/auth/linuxdo | jq . 2>/dev/null || curl -s http://localhost:12322/api/auth/linuxdo
echo ""
echo ""

echo "=== 修复完成！==="
echo "请访问 https://codex.zenscaleai.com/login 测试LinuxDo登录"
