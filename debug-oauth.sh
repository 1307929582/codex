#!/bin/bash
echo "=== 调试LinuxDo OAuth配置 ==="
echo ""
echo "1. 数据库中的配置："
docker exec codex-postgres psql -U postgres -d codex_gateway -c "SELECT id, linuxdo_client_id, linuxdo_client_secret, linuxdo_enabled FROM system_settings;"
echo ""
echo "2. 表结构："
docker exec codex-postgres psql -U postgres -d codex_gateway -c "\d system_settings" | grep linuxdo
echo ""
echo "3. 测试OAuth端点："
curl http://localhost:12322/api/auth/linuxdo
echo ""
