# 🚀 Codex Gateway 快速参考

一页纸快速参考指南，包含所有常用命令和配置。

---

## 📦 一键部署

```bash
git clone https://github.com/1307929582/codex.git
cd codex
./deploy-auto.sh
```

访问: `http://localhost:3000`

---

## 🔗 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 前端 | http://localhost:3000 | 用户界面 |
| 管理员面板 | http://localhost:3000/admin | 管理功能 |
| 后端API | http://localhost:8080 | API服务 |
| 健康检查 | http://localhost:8080/health | 服务状态 |

---

## 🐳 Docker命令

### 基本操作
```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启服务
docker-compose restart

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres
```

### 重新构建
```bash
# 重新构建所有镜像
docker-compose build

# 重新构建特定服务
docker-compose build backend

# 无缓存构建
docker-compose build --no-cache

# 构建并启动
docker-compose up -d --build
```

### 清理
```bash
# 停止并删除容器
docker-compose down

# 停止并删除容器和卷（删除数据）
docker-compose down -v

# 删除所有未使用的镜像
docker image prune -a
```

---

## 🗄️ 数据库操作

### 连接数据库
```bash
# 进入PostgreSQL容器
docker exec -it codex-postgres psql -U postgres -d codex_gateway

# 或使用一行命令
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c "SELECT * FROM users;"
```

### 常用SQL命令
```sql
-- 查看所有用户
SELECT id, email, balance, status, role FROM users;

-- 提升用户为管理员
UPDATE users SET role = 'admin' WHERE email = 'user@example.com';

-- 查看用户余额
SELECT email, balance FROM users ORDER BY balance DESC;

-- 查看系统设置
SELECT * FROM system_settings;

-- 查看操作日志
SELECT * FROM admin_logs ORDER BY created_at DESC LIMIT 10;

-- 查看使用记录
SELECT * FROM usage_records ORDER BY created_at DESC LIMIT 10;

-- 退出
\q
```

### 数据库备份与恢复
```bash
# 备份数据库
docker exec codex-postgres pg_dump -U postgres codex_gateway > backup_$(date +%Y%m%d).sql

# 恢复数据库
docker exec -i codex-postgres psql -U postgres -d codex_gateway < backup_20260119.sql

# 备份到容器内
docker exec codex-postgres pg_dump -U postgres codex_gateway > /tmp/backup.sql

# 从容器复制备份文件
docker cp codex-postgres:/tmp/backup.sql ./backup.sql
```

---

## 👤 用户管理

### 创建管理员
```bash
# 方式1: 通过SQL
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';"

# 方式2: 交互式
docker exec -it codex-postgres psql -U postgres -d codex_gateway
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
\q
```

### 查看用户信息
```bash
# 查看所有用户
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT email, balance, status, role FROM users;"

# 查看特定用户
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT * FROM users WHERE email = 'user@example.com';"
```

### 调整用户余额
```bash
# 充值100元
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "UPDATE users SET balance = balance + 100 WHERE email = 'user@example.com';"

# 扣除50元
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "UPDATE users SET balance = balance - 50 WHERE email = 'user@example.com';"
```

---

## 🔑 API使用

### 用户认证
```bash
# 注册
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# 响应包含token
# {"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

### API密钥管理
```bash
# 创建API密钥（需要JWT token）
curl -X POST http://localhost:8080/api/keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My API Key"}'

# 列出API密钥
curl http://localhost:8080/api/keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 删除API密钥
curl -X DELETE http://localhost:8080/api/keys/KEY_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### OpenAI代理调用
```bash
# 使用API密钥调用
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

---

## ⚙️ 配置管理

### 环境变量（.env）
```bash
# 必需的环境变量（仅3个）
DB_PASSWORD=your-secure-password
JWT_SECRET=your-jwt-secret-min-32-chars
NEXT_PUBLIC_API_URL=http://localhost:8080

# 可选的环境变量（使用默认值）
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_NAME=codex_gateway
SERVER_PORT=8080
```

### 系统设置（管理员面板）
访问 `http://localhost:3000/admin/settings` 配置：

- **OpenAI API密钥**: 在Web界面配置
- **OpenAI Base URL**: 支持自定义代理
- **系统公告**: 显示给所有用户
- **新用户默认余额**: 注册时赠送
- **最小充值金额**: 充值限制
- **注册开关**: 是否允许注册

---

## 🔍 故障排查

### 检查服务状态
```bash
# 查看所有容器
docker-compose ps

# 查看容器资源使用
docker stats

# 测试后端健康
curl http://localhost:8080/health

# 测试前端访问
curl -I http://localhost:3000
```

### 查看日志
```bash
# 实时查看所有日志
docker-compose logs -f

# 查看最近100行
docker-compose logs --tail=100

# 查看特定服务
docker-compose logs -f backend

# 查看错误日志
docker-compose logs | grep -i error
```

### 重启服务
```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart backend
docker-compose restart frontend
docker-compose restart postgres

# 完全重新部署
docker-compose down
docker-compose up -d --build
```

### 清除并重新开始
```bash
# 停止并删除所有容器和数据
docker-compose down -v

# 删除.env文件
rm .env

# 重新部署
./deploy-auto.sh
```

---

## 📊 监控命令

### 系统资源
```bash
# 查看容器资源使用
docker stats

# 查看磁盘使用
df -h

# 查看Docker磁盘使用
docker system df
```

### 数据库统计
```bash
# 用户数量
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT COUNT(*) FROM users;"

# 今日注册用户
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURRENT_DATE;"

# 总余额
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT SUM(balance) FROM users;"

# API密钥数量
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT COUNT(*) FROM api_keys;"

# 今日使用记录
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT COUNT(*), SUM(cost) FROM usage_records WHERE DATE(created_at) = CURRENT_DATE;"
```

---

## 🔐 安全操作

### 修改密码
```bash
# 生成新的JWT密钥
openssl rand -base64 32

# 生成新的数据库密码
openssl rand -base64 24

# 更新.env文件
nano .env

# 重启服务
docker-compose restart
```

### 查看操作日志
```bash
# 查看最近的管理员操作
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT * FROM admin_logs ORDER BY created_at DESC LIMIT 10;"

# 查看特定管理员的操作
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT * FROM admin_logs WHERE admin_id = 'USER_ID' ORDER BY created_at DESC;"
```

---

## 📁 文件位置

### 重要文件
```
codex中转/
├── .env                    # 环境变量配置
├── docker-compose.yml      # Docker配置
├── deploy-auto.sh          # 一键部署脚本
├── README.md              # 项目文档
├── ADMIN_GUIDE.md         # 管理员指南
└── FEATURES_DEMO.md       # 功能演示
```

### 日志文件
```bash
# Docker日志
docker-compose logs > logs.txt

# 后端日志
docker-compose logs backend > backend.log

# 前端日志
docker-compose logs frontend > frontend.log

# 数据库日志
docker-compose logs postgres > postgres.log
```

---

## 🆘 快速帮助

### 常见问题

**Q: 如何重置管理员密码？**
```bash
# 1. 连接数据库
docker exec -it codex-postgres psql -U postgres -d codex_gateway

# 2. 生成新密码哈希（使用bcrypt）
# 在Go中: bcrypt.GenerateFromPassword([]byte("newpassword"), bcrypt.DefaultCost)

# 3. 更新密码
UPDATE users SET password_hash = 'NEW_HASH' WHERE email = 'admin@example.com';
```

**Q: 如何查看当前OpenAI配置？**
```bash
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "SELECT openai_base_url FROM system_settings;"
```

**Q: 如何清空所有数据？**
```bash
docker-compose down -v
./deploy-auto.sh
```

### 获取更多帮助
- 📚 完整文档: [README_DEPLOY.md](./README_DEPLOY.md)
- 🛡️ 管理员指南: [ADMIN_GUIDE.md](./ADMIN_GUIDE.md)
- 🎬 功能演示: [FEATURES_DEMO.md](./FEATURES_DEMO.md)
- 🐛 问题反馈: https://github.com/1307929582/codex/issues

---

**版本**: v2.0.0
**最后更新**: 2026-01-19
**打印友好**: 是
