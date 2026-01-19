# Codex Gateway

商业级OpenAI API网关，支持多用户、计费、管理员面板等企业级功能。

## ✨ 最新版本

**v2.0.0 (2026-01-19)**:
- 🎉 WordPress风格的首次安装向导
- 🎯 零配置一键部署
- 🛡️ 完整的管理员面板
- 🔐 所有配置移至Web界面

**v1.0.1 (2026-01-19)**: 修复了多个关键安全漏洞，包括DoS攻击、财务漏洞和JWT安全问题。详见 [SECURITY_FIXES.md](./SECURITY_FIXES.md)

## 项目结构

```
.
├── cmd/gateway/           # 后端主程序
├── internal/              # 后端核心代码
│   ├── handlers/         # API处理器
│   ├── middleware/       # 中间件
│   ├── models/           # 数据模型
│   ├── database/         # 数据库
│   └── config/           # 配置
├── frontend/              # Next.js前端
└── docs/                  # 文档
```

## 功能特性

### 🎯 零配置部署
- ✅ 一键部署脚本（自动生成所有配置）
- ✅ WordPress风格的首次安装向导
- ✅ Web界面完成所有配置
- ✅ 无需手动编辑配置文件
- ✅ 无需执行SQL命令

### 🛡️ 管理员面板
- ✅ 用户管理（查看、暂停、调整余额）
- ✅ 系统设置（公告、默认余额、注册开关）
- ✅ OpenAI配置（API密钥、Base URL）
- ✅ 统计分析（用户数、收入、消费、API使用）
- ✅ 操作审计（所有管理员操作记录）

### 💰 计费系统
- ✅ 原子化计费（事务安全）
- ✅ 多模型定价支持
- ✅ 实时余额扣除
- ✅ 使用量统计
- ✅ 账户余额管理

### 🔐 安全特性
- ✅ JWT用户认证
- ✅ API密钥管理
- ✅ 角色权限控制（user/admin/super_admin）
- ✅ 请求限流保护
- ✅ 操作审计日志

### 🚀 API代理
- ✅ OpenAI API完整代理
- ✅ 自动计费和扣费
- ✅ 请求日志记录
- ✅ 错误处理和重试

### 📊 前端功能
- ✅ 用户登录/注册
- ✅ Dashboard仪表盘
- ✅ API密钥CRUD
- ✅ 使用记录查询
- ✅ 账户管理
- ✅ 管理员面板

## 🚀 快速开始（零配置）

### 一键部署（推荐）

```bash
# 1. 克隆代码
git clone https://github.com/1307929582/codex.git
cd codex

# 2. 一键部署
./deploy-auto.sh

# 3. 打开浏览器
open http://localhost:3000
```

就这么简单！系统会自动：
- ✅ 生成安全的数据库密码和JWT密钥
- ✅ 构建并启动所有服务
- ✅ 跳转到首次安装向导
- ✅ 在Web界面完成配置

### 首次安装向导

访问 `http://localhost:3000` 后，系统会自动引导您完成3步配置：

1. **创建管理员账户**
   - 输入邮箱和密码

2. **配置OpenAI**
   - 输入OpenAI API密钥
   - 可选：自定义Base URL

3. **系统设置**
   - 可选：设置系统公告
   - 可选：配置新用户默认余额
   - 可选：是否允许注册

完成后自动登录到管理员面板！

### 访问地址

- **前端**: http://localhost:3000
- **管理员面板**: http://localhost:3000/admin
- **后端API**: http://localhost:8080

---

## 📚 详细文档

### 快速导航
- 📖 [DOCUMENTATION_INDEX.md](./DOCUMENTATION_INDEX.md) - **完整文档索引**（推荐从这里开始）
- 🚀 [QUICK_START.md](./QUICK_START.md) - 3个命令快速部署
- 📋 [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - 常用命令快速参考

### 核心文档
- [README_DEPLOY.md](./README_DEPLOY.md) - 完整部署指南
- [ADMIN_GUIDE.md](./ADMIN_GUIDE.md) - 管理员面板使用指南
- [FEATURES_DEMO.md](./FEATURES_DEMO.md) - 功能演示（带界面）
- [API_DOCUMENTATION.md](./API_DOCUMENTATION.md) - API文档

### 运维文档
- [DEPLOYMENT_CHECKLIST.md](./DEPLOYMENT_CHECKLIST.md) - 部署检查清单
- [DEPLOYMENT_FINAL.md](./DEPLOYMENT_FINAL.md) - 最终部署文档

### 项目文档
- [PROJECT_SUMMARY.md](./PROJECT_SUMMARY.md) - 项目技术总结
- [PROJECT_COMPLETION_REPORT.md](./PROJECT_COMPLETION_REPORT.md) - 项目完成报告
- [RELEASE_NOTES_v2.0.md](./RELEASE_NOTES_v2.0.md) - v2.0发布说明
- [SECURITY_FIXES.md](./SECURITY_FIXES.md) - 安全更新日志

---

## 🛠️ 手动部署（开发环境）

如果您需要手动部署或进行开发，请参考以下步骤：

### 1. 后端

```bash
# 安装依赖
go mod download

# 配置环境变量（最小配置）
cat > .env <<EOF
DB_PASSWORD=your-password
JWT_SECRET=$(openssl rand -base64 32)
EOF

# 启动数据库
docker-compose up -d postgres

# 启动服务
go run cmd/gateway/main.go
```

### 2. 前端

```bash
cd frontend

# 安装依赖
npm install

# 配置环境变量
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local

# 启动开发服务器
npm run dev
```

---

## 🐳 Docker部署

### 使用Docker Compose（推荐）

```bash
# 一键部署
./deploy-auto.sh

# 或手动部署
docker-compose up -d
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看后端日志
docker-compose logs -f backend

# 查看前端日志
docker-compose logs -f frontend
```

### 停止服务

```bash
docker-compose down
```

---

## 🔧 管理员操作

### 创建管理员（如果跳过了安装向导）

```bash
# 方式1：通过数据库
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';"

# 方式2：通过SQL客户端
docker exec -it codex-postgres psql -U postgres -d codex_gateway
UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';
\q
```

### 管理员面板功能

访问 `http://localhost:3000/admin` 可以：

- 📊 查看系统统计（用户数、收入、消费）
- 👥 管理用户（查看、暂停、调整余额）
- ⚙️ 配置系统（OpenAI、公告、注册开关）
- 📝 查看操作日志（审计追踪）

详见 [ADMIN_GUIDE.md](./ADMIN_GUIDE.md)

---

## 🔐 环境变量说明

### 必需的环境变量（仅3个）

```bash
# 数据库密码
DB_PASSWORD=your-secure-password

# JWT密钥（至少32字符）
JWT_SECRET=your-jwt-secret-min-32-chars

# 前端API地址
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 可选的环境变量

```bash
# 数据库配置（使用默认值）
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_NAME=codex_gateway

# 服务端口
SERVER_PORT=8080
```

### ❌ 不再需要的环境变量

```bash
# 以下配置已移到管理员面板
OPENAI_API_KEY=xxx  # ❌ 在管理员面板配置
OPENAI_BASE_URL=xxx # ❌ 在管理员面板配置
```

---

## 技术栈

**后端**:
- Go 1.21+
- Gin (HTTP框架)
- GORM (ORM)
- PostgreSQL
- JWT认证

**前端**:
- Next.js 14
- TypeScript
- Tailwind CSS
- TanStack Query
- Zustand

## License

MIT
