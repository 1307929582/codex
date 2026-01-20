#!/bin/bash

echo "=== Zenscale Codex - 套餐+支付系统部署脚本 ==="
echo ""
echo "此脚本将部署完整的套餐购买和Credit支付系统"
echo ""

# 1. 拉取最新代码
echo "1. 拉取最新代码..."
git pull

# 2. 运行数据库迁移
echo ""
echo "2. 运行数据库迁移..."
docker exec codex-gateway-db-1 psql -U postgres -d codex_gateway -f /app/migrations/add_packages_and_payment.sql

# 3. 重新构建并启动服务
echo ""
echo "3. 重新构建并启动服务..."
docker compose up -d --build

# 4. 等待服务启动
echo ""
echo "4. 等待服务启动..."
sleep 10

# 5. 检查服务状态
echo ""
echo "5. 检查服务状态..."
docker compose ps

echo ""
echo "=== 部署完成 ==="
echo ""
echo "新功能："
echo "  - 管理员可以在 /admin/packages 管理套餐"
echo "  - 管理员可以在 /admin/settings 配置Credit支付"
echo "  - 用户可以在 /packages 购买套餐"
echo "  - 用户可以在 /dashboard 查看套餐状态和每日使用情况"
echo ""
echo "配置步骤："
echo "  1. 登录管理员账号"
echo "  2. 进入系统设置，配置Credit支付参数："
echo "     - Credit PID (Client ID)"
echo "     - Credit Key (Client Secret)"
echo "     - Notify URL: https://your-domain.com/api/payment/credit/notify"
echo "     - Return URL: https://your-domain.com/packages"
echo "  3. 启用Credit支付"
echo "  4. 在套餐管理中创建或编辑套餐"
echo ""
echo "注意事项："
echo "  - 确保Credit回调URL可以被外网访问"
echo "  - 套餐价格单位为美元"
echo "  - 每日限额在UTC+8时区的0点重置"
echo "  - 用户有活跃套餐时优先使用套餐额度"
echo ""
