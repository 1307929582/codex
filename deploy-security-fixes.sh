#!/bin/bash

echo "=== 套餐+支付系统 - 安全修复部署脚本 ==="
echo ""
echo "本次更新包含关键安全和并发修复"
echo ""

# 1. 拉取最新代码
echo "1. 拉取最新代码..."
git pull

# 2. 停止现有服务
echo ""
echo "2. 停止现有服务..."
docker compose down

# 3. 运行数据库迁移（如果还没运行过）
echo ""
echo "3. 检查数据库迁移..."
docker compose up -d postgres
sleep 5
docker exec codex-postgres psql -U postgres -d codex_gateway -c "\dt packages" > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "运行数据库迁移..."
    docker cp migrations/add_packages_and_payment.sql codex-postgres:/tmp/
    docker exec codex-postgres psql -U postgres -d codex_gateway -f /tmp/add_packages_and_payment.sql
else
    echo "数据库表已存在，跳过迁移"
fi

# 4. 重新构建并启动所有服务
echo ""
echo "4. 重新构建并启动服务..."
docker compose up -d --build

# 5. 等待服务启动
echo ""
echo "5. 等待服务启动..."
sleep 15

# 6. 检查服务状态
echo ""
echo "6. 检查服务状态..."
docker compose ps

# 7. 检查网络连接
echo ""
echo "7. 检查容器网络..."
docker network inspect codex-中转_codex-network 2>/dev/null || docker network inspect codex_codex-network 2>/dev/null || echo "网络检查跳过"

# 8. 测试后端健康
echo ""
echo "8. 测试后端健康..."
curl -s http://localhost:12322/health || echo "后端健康检查失败"

# 9. 测试前端
echo ""
echo "9. 测试前端..."
curl -s http://localhost:12321 > /dev/null && echo "前端正常" || echo "前端检查失败"

# 10. 查看日志
echo ""
echo "10. 查看最近的日志..."
echo "--- Backend logs ---"
docker compose logs backend --tail=20
echo ""
echo "--- Frontend logs ---"
docker compose logs frontend --tail=20

echo ""
echo "=== 部署完成 ==="
echo ""
echo "修复内容："
echo "  ✅ 修复余额扣费竞态条件（使用原子操作）"
echo "  ✅ 修复每日使用量并发问题（原子更新）"
echo "  ✅ 增强支付回调安全性（订单过期检查、详细日志）"
echo "  ✅ 修复前端网络连接问题（Docker网络配置）"
echo ""
echo "下一步："
echo "  1. 访问 http://localhost:12321 测试前端"
echo "  2. 登录管理员账号配置Credit支付"
echo "  3. 创建测试套餐"
echo "  4. 测试购买流程"
echo ""
echo "如果遇到问题："
echo "  - 查看日志: docker compose logs -f [backend|frontend]"
echo "  - 检查网络: docker network ls"
echo "  - 重启服务: docker compose restart"
echo ""
