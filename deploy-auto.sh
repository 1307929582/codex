#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "================================"
echo "  Codex Gateway 一键部署"
echo "================================"
echo ""

# Check requirements
echo -e "${GREEN}[1/7]${NC} 检查系统要求..."
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: 未安装Docker${NC}"
    exit 1
fi

# Check for docker-compose (V1) or docker compose (V2)
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    echo -e "${RED}错误: 未安装Docker Compose${NC}"
    exit 1
fi

echo -e "${GREEN}✓${NC} Docker已安装"
echo -e "${GREEN}✓${NC} Docker Compose已安装 ($DOCKER_COMPOSE)"

# Stop existing services
echo ""
echo -e "${GREEN}[2/7]${NC} 停止现有服务..."
$DOCKER_COMPOSE down 2>/dev/null || true

# Generate .env if not exists
echo ""
echo -e "${GREEN}[3/7]${NC} 生成配置文件..."

if [ ! -f .env ]; then
    echo -e "${YELLOW}生成新的.env文件...${NC}"

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
SERVER_PORT=8080

# Frontend Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
EOF

    echo -e "${GREEN}✓${NC} 配置文件已生成"
    echo -e "${YELLOW}注意: 数据库密码和JWT密钥已自动生成${NC}"
else
    echo -e "${GREEN}✓${NC} 使用现有配置文件"
fi

# Build images
echo ""
echo -e "${GREEN}[4/7]${NC} 构建Docker镜像..."
$DOCKER_COMPOSE build

# Start services
echo ""
echo -e "${GREEN}[5/7]${NC} 启动服务..."
$DOCKER_COMPOSE up -d

# Wait for services
echo ""
echo -e "${GREEN}[6/7]${NC} 等待服务就绪..."
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
    if curl -s http://localhost:8080/health &>/dev/null; then
        echo -e " ${GREEN}✓${NC}"
        break
    fi
    echo -n "."
    sleep 1
done

# Check if admin exists
echo ""
echo -e "${GREEN}[7/7]${NC} 检查管理员账户..."

ADMIN_COUNT=$(docker exec codex-postgres psql -U postgres -d codex_gateway -t -c "SELECT COUNT(*) FROM users WHERE role IN ('admin', 'super_admin');" 2>/dev/null | tr -d ' ' || echo "0")

if [ "$ADMIN_COUNT" = "0" ]; then
    echo -e "${YELLOW}未找到管理员账户${NC}"
    echo ""
    echo "请选择创建管理员的方式："
    echo "1) 现在注册新账户并提升为管理员"
    echo "2) 稍后手动创建"
    echo ""
    read -p "请选择 [1/2]: " choice

    if [ "$choice" = "1" ]; then
        echo ""
        read -p "请输入管理员邮箱: " admin_email
        read -s -p "请输入密码: " admin_password
        echo ""

        # Register user
        echo -n "注册账户..."
        REGISTER_RESPONSE=$(curl -s -X POST http://localhost:12322/api/auth/register \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"$admin_email\",\"password\":\"$admin_password\"}")

        if echo "$REGISTER_RESPONSE" | grep -q "token"; then
            echo -e " ${GREEN}✓${NC}"

            # Promote to admin
            echo -n "提升为管理员..."
            docker exec codex-postgres psql -U postgres -d codex_gateway -c \
                "UPDATE users SET role = 'admin' WHERE email = '$admin_email';" &>/dev/null
            echo -e " ${GREEN}✓${NC}"

            echo ""
            echo -e "${GREEN}✓ 管理员账户创建成功！${NC}"
            echo ""
            echo "管理员邮箱: $admin_email"
        else
            echo -e " ${RED}✗${NC}"
            echo -e "${RED}注册失败，请稍后手动创建${NC}"
        fi
    fi
else
    echo -e "${GREEN}✓${NC} 已存在 $ADMIN_COUNT 个管理员账户"
fi

# Show status
echo ""
echo "================================"
echo -e "${GREEN}部署完成！${NC}"
echo "================================"
echo ""
echo "访问地址："
echo "  前端:        http://localhost:12321"
echo "  管理员面板:  http://localhost:12321/admin"
echo "  后端API:     http://localhost:12322"
echo ""
echo "下一步："
echo "  1. 访问 http://localhost:12321/admin"
echo "  2. 使用管理员账户登录"
echo "  3. 在 系统设置 中配置 OpenAI API 密钥"
echo ""
echo "查看日志："
echo "  $DOCKER_COMPOSE logs -f"
echo ""
echo "停止服务："
echo "  $DOCKER_COMPOSE down"
echo ""
