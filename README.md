# Codex Gateway

商业级OpenAI API网关，支持多用户、计费、限流等功能。

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

### 后端
- ✅ JWT用户认证
- ✅ API密钥管理
- ✅ OpenAI代理服务
- ✅ 原子化计费系统
- ✅ 使用量统计
- ✅ 账户余额管理

### 前端
- ✅ 用户登录/注册
- ✅ Dashboard仪表盘
- ✅ API密钥CRUD
- ✅ 使用记录查询
- ✅ 账户管理

## 快速开始

### 1. 后端

```bash
# 安装依赖
go mod download

# 配置环境变量
cp .env.example .env
# 编辑 .env

# 创建数据库
createdb codex_gateway

# 启动服务
go run cmd/gateway/main.go
```

### 2. 前端

```bash
cd frontend

# 安装依赖
npm install

# 配置环境变量
cp .env.local .env.local
# 编辑 .env.local

# 启动开发服务器
npm run dev
```

## 生产部署

参见 [DEPLOYMENT.md](./DEPLOYMENT.md)

## API文档

参见 [API_DOCUMENTATION.md](./API_DOCUMENTATION.md)

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
