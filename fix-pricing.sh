#!/bin/bash

# Codex Gateway 定价修复脚本
# 用途：重置数据库中的模型定价为正确值

set -e

echo "========================================="
echo "Codex Gateway 定价修复脚本"
echo "========================================="
echo ""

# 检查是否提供了管理员 token
if [ -z "$ADMIN_TOKEN" ]; then
    echo "错误：请设置 ADMIN_TOKEN 环境变量"
    echo ""
    echo "使用方法："
    echo "  export ADMIN_TOKEN='your-admin-token'"
    echo "  ./fix-pricing.sh"
    echo ""
    echo "或者："
    echo "  ADMIN_TOKEN='your-admin-token' ./fix-pricing.sh"
    echo ""
    exit 1
fi

# 服务器地址
SERVER="http://23.80.88.63:12321"

echo "步骤 1: 检查服务器连接..."
if curl -s -f "$SERVER/health" > /dev/null; then
    echo "✓ 服务器连接正常"
else
    echo "✗ 无法连接到服务器"
    exit 1
fi

echo ""
echo "步骤 2: 重置定价数据..."
RESPONSE=$(curl -s -X POST "$SERVER/api/admin/pricing/reset" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json")

# 检查响应
if echo "$RESPONSE" | grep -q "Pricing reset successfully"; then
    echo "✓ 定价重置成功"
    echo ""
    echo "新的定价数据："
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
else
    echo "✗ 定价重置失败"
    echo "响应："
    echo "$RESPONSE"
    exit 1
fi

echo ""
echo "========================================="
echo "���复完成！"
echo "========================================="
echo ""
echo "现在您可以测试 API 调用，费用应该恢复正常。"
echo ""
