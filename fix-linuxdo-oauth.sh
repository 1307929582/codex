#!/bin/bash
# LinuxDo OAuth 503错误诊断和修复脚本

echo "========================================="
echo "LinuxDo OAuth 503错误诊断"
echo "========================================="
echo ""

# 1. 检查后端日志
echo "1. 检查后端日志中的错误..."
echo "-----------------------------------"
docker-compose logs backend 2>&1 | tail -50 | grep -i "error\|linuxdo\|oauth\|503" || echo "未发现明显错误"
echo ""

# 2. 检查数据库表结构
echo "2. 检查system_settings表结构..."
echo "-----------------------------------"
docker exec codex-gateway-db-1 psql -U codex_user -d codex_gateway -c "\d system_settings" 2>&1
echo ""

# 3. 检查LinuxDo配置数据
echo "3. 检查LinuxDo OAuth配置..."
echo "-----------------------------------"
docker exec codex-gateway-db-1 psql -U codex_user -d codex_gateway -c "SELECT id, linuxdo_client_id, linuxdo_enabled FROM system_settings;" 2>&1
echo ""

# 4. 检查迁移日志
echo "4. 检查数据��迁移日志..."
echo "-----------------------------------"
docker-compose logs backend 2>&1 | grep -i "migration" | tail -20
echo ""

echo "========================================="
echo "开始修复"
echo "========================================="
echo ""

# 5. 添加缺失的字段
echo "5. 添加LinuxDo OAuth字段..."
echo "-----------------------------------"
docker exec codex-gateway-db-1 psql -U codex_user -d codex_gateway << 'SQL'
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_id VARCHAR(255);
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_secret VARCHAR(255);
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_enabled BOOLEAN DEFAULT false;
SQL
echo "字段添加完成"
echo ""

# 6. 添加OAuth用户字段
echo "6. 添加用户OAuth字段..."
echo "-----------------------------------"
docker exec codex-gateway-db-1 psql -U codex_user -d codex_gateway << 'SQL'
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
CREATE INDEX IF NOT EXISTS idx_oauth ON users(oauth_provider, oauth_id);
UPDATE users SET oauth_provider = 'email' WHERE oauth_provider IS NULL;
SQL
echo "用户OAuth字段添加完成"
echo ""

# 7. 插入/更新LinuxDo配置
echo "7. 配置LinuxDo OAuth..."
echo "-----------------------------------"
docker exec codex-gateway-db-1 psql -U codex_user -d codex_gateway << 'SQL'
INSERT INTO system_settings (id, announcement, default_balance, min_recharge_amount, registration_enabled, linuxdo_client_id, linuxdo_client_secret, linuxdo_enabled)
VALUES (1, '', 0, 10, true, 'kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ', 'XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg', true)
ON CONFLICT (id) DO UPDATE SET
  linuxdo_client_id = 'kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ',
  linuxdo_client_secret = 'XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg',
  linuxdo_enabled = true;
SQL
echo "LinuxDo OAuth配置完成"
echo ""

# 8. 重启后端服务
echo "8. 重启后端服务..."
echo "-----------------------------------"
docker-compose restart backend
echo "等待服务启动..."
sleep 10
echo ""

# 9. 验证配置
echo "9. 验证配置..."
echo "-----------------------------------"
docker exec codex-gateway-db-1 psql -U codex_user -d codex_gateway -c "SELECT id, linuxdo_client_id, linuxdo_enabled FROM system_settings;"
echo ""

# 10. 测试OAuth端点
echo "10. 测试LinuxDo OAuth端点..."
echo "-----------------------------------"
curl -s http://localhost:12322/api/auth/linuxdo | jq . || curl -s http://localhost:12322/api/auth/linuxdo
echo ""
echo ""

echo "========================================="
echo "修复完成！"
echo "========================================="
echo ""
echo "请访问 https://codex.zenscaleai.com/login 测试LinuxDo登录"
echo ""
