#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "================================"
echo "  Codex Gateway 部署修复工具"
echo "================================"
echo ""

# 1. 停止所有服务
echo -e "${GREEN}[1/6]${NC} 停止现有服务..."
docker compose down 2>/dev/null || true

# 2. 删除旧的数据库数据
echo ""
echo -e "${GREEN}[2/6]${NC} 清理旧数据..."
docker volume rm codex_postgres_data 2>/dev/null || true

# 3. 生成新的.env文件
echo ""
echo -e "${GREEN}[3/6]${NC} 生成配置文件..."

# Generate random passwords
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
JWT_SECRET=$(openssl rand -base64 48 | tr -d "=+/" | cut -c1-40)

cat > .env <<EOF
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=codex_gateway
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=${JWT_SECRET}

# Server Configuration
SERVER_PORT=12322

# Frontend Configuration
NEXT_PUBLIC_API_URL=http://localhost:12322
EOF

echo -e "${GREEN}✓${NC} 配置文件已生成"
echo -e "${YELLOW}数据库密码: ${DB_PASSWORD}${NC}"
echo -e "${YELLOW}JWT密钥: ${JWT_SECRET}${NC}"

# 4. 验证.env文件
echo ""
echo -e "${GREEN}[4/6]${NC} 验证配置文件..."
if [ -f .env ]; then
    echo -e "${GREEN}✓${NC} .env 文件存在"
    if grep -q "DB_PASSWORD=" .env && [ -n "$(grep DB_PASSWORD= .env | cut -d= -f2)" ]; then
        echo -e "${GREEN}✓${NC} DB_PASSWORD 已设置"
    else
        echo -e "${RED}✗${NC} DB_PASSWORD 未设置或为空"
        exit 1
    fi
else
    echo -e "${RED}✗${NC} .env 文件不存在"
    exit 1
fi

# 5. 构建并启动服务
echo ""
echo -e "${GREEN}[5/6]${NC} 构建并启动服务..."
docker compose build
docker compose up -d

# 6. 等待服务就绪
echo ""
echo -e "${GREEN}[6/6]${NC} 等待服务就绪..."

echo -n "等待数据库启动"
for i in {1..30}; do
    if docker exec codex-postgres pg_isready -U postgres &>/dev/null; then
        echo -e " ${GREEN}✓${NC}"
        break
    fi
    echo -n "."
    sleep 1
done

echo -n "等待后端启动"
for i in {1..30}; do
    if curl -s http://localhost:12322/health &>/dev/null; then
        echo -e " ${GREEN}✓${NC}"
        break
    fi
    echo -n "."
    sleep 1
done

# 显示状态
echo ""
echo "================================"
echo -e "${GREEN}修复完成！${NC}"
echo "================================"
echo ""
echo "访问地址："
echo "  前端:        http://localhost:12321"
echo "  安装向导:    http://localhost:12321/setup"
echo "  后端API:     http://localhost:12322"
echo ""
echo "下一步："
echo "  1. 访问 http://localhost:12321/setup"
echo "  2. 按照向导创建管理员账户"
echo "  3. 配置 OpenAI API 密钥"
echo ""
echo "查看日志："
echo "  docker compose logs -f"
echo ""
