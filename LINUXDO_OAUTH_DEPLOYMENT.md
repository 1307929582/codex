# 🚀 LinuxDo OAuth 登录集成 - 部署指南

## ✅ 已完成的功能

### 后端实现
- ✅ 扩展User模型，添加OAuth字段（oauth_provider, oauth_id, username, avatar_url）
- ✅ 实现LinuxDo OAuth 2.0授权流程
- ✅ 支持账户关联（通过邮箱）
- ✅ 密码字段改为可选（OAuth用户无需密码）
- ✅ 数据库迁移文件
- ✅ OAuth配置管理

### 前端实现
- ✅ 登录页面添加"使用LinuxDo登录"按钮
- ✅ OAuth回调处理页面
- ✅ 自动获取用户信息并登录

### 安全特性
- ✅ CSRF保护（state参数）
- ✅ 安全的Cookie处理
- ✅ JWT token生成
- ✅ 7天token有效期

---

## 📋 LinuxDo OAuth配置

### 应用信息
- **Client ID**: `kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ`
- **Client Secret**: `XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg`
- **应用名**: codex
- **应用主页**: https://codex.zenscaleai.com/
- **回调地址**: https://codex.zenscaleai.com/api/auth/linuxdo/callback

---

## 🔧 部署步骤

### 1. 更新环境变量

在服务器上编辑 `.env` 文件，添加以下配置：

```bash
ssh root@23.80.88.63
cd /root/codex-gateway
nano .env
```

添加以下内容：

```env
# LinuxDo OAuth Configuration
LINUXDO_CLIENT_ID=kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ
LINUXDO_CLIENT_SECRET=XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg
LINUXDO_REDIRECT_URL=https://codex.zenscaleai.com/api/auth/linuxdo/callback
FRONTEND_URL=https://codex.zenscaleai.com
DEFAULT_BALANCE=0
```

保存并退出（Ctrl+X, Y, Enter）

### 2. 运行数据库迁移

```bash
# 连接到数据库
docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway

# 执行迁移SQL
\i /path/to/migrations/002_add_oauth_fields.sql

# 或者手动执行：
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
CREATE INDEX IF NOT EXISTS idx_oauth ON users(oauth_provider, oauth_id);
UPDATE users SET oauth_provider = 'email' WHERE oauth_provider IS NULL;

# 退出
\q
```

### 3. 部署新代码

```bash
cd /root/codex-gateway
git pull origin main
./deploy-auto.sh
```

### 4. 验证部署

检查服务是否正常启动：

```bash
docker-compose logs -f backend | grep -i "oauth\|linuxdo"
```

应该看到类似的日志：
```
Server starting on port 12322
```

---

## 🧪 测试OAuth登录

### 1. 访问登录页面

打开浏览器访问：https://codex.zenscaleai.com/login

### 2. 点击"使用LinuxDo登录"按钮

应该会跳转到LinuxDo授权页面：
```
https://linux.do/oauth2/authorize?client_id=...&redirect_uri=...&response_type=code&scope=read&state=...
```

### 3. 授权并登录

- 在LinuxDo页面点击"授权"
- 自动跳转回您的网站
- 自动登录并进入Dashboard

### 4. 验证用户信息

登录后，检查用户信息是否正确：
- 用户名应该显示LinuxDo用户名
- 头像应该显示LinuxDo头像
- 邮箱应该是LinuxDo邮箱（如果提供）

---

## 🔍 OAuth登录流程

```
用户点击"使用LinuxDo登录"
    ↓
GET /api/auth/linuxdo
    ↓
返回LinuxDo授权URL
    ↓
跳转到LinuxDo授权页面
    ↓
用户授权
    ↓
LinuxDo回调: GET /api/auth/linuxdo/callback?code=xxx&state=xxx
    ↓
验证state（CSRF保护）
    ↓
用code换取access_token
    ↓
使用access_token获取用户信息
    ↓
查找或创建用户账户
    ↓
生成JWT token
    ↓
重定向到前端: /auth/callback?token=xxx
    ↓
前端获取用户信息
    ↓
登录成功，跳转到Dashboard
```

---

## 📊 账户关联策略

### 场景1：新用户（LinuxDo首次登录）
- 创建新账户
- oauth_provider = "linuxdo"
- oauth_id = LinuxDo用户ID
- email = LinuxDo邮箱（或生成占位邮箱）
- 初始余额 = DEFAULT_BALANCE

### 场景2：已有邮箱账户
- 如果LinuxDo邮箱与现有账户匹配
- 关联OAuth信息到现有账户
- 更新oauth_provider和oauth_id
- 保留原有余额和数据

### 场景3：已有LinuxDo账户
- 直接登录
- 更新用户名和头像
- 保留所有数据

---

## 🐛 故障排查

### 问题1：点击LinuxDo登录无反应

**检查**：
```bash
docker-compose logs backend | grep "linuxdo"
```

**可能原因**：
- 环境变量未设置
- OAuth配置错误

**解决方案**：
```bash
# 检查环境变量
docker-compose exec backend env | grep LINUXDO
```

### 问题2：授权后回调失败

**检查回调URL**：
- 确认LinuxDo应用配置中的回调URL正确
- 应该是：`https://codex.zenscaleai.com/api/auth/linuxdo/callback`

**检查日志**：
```bash
docker-compose logs backend | grep "callback"
```

### 问题3：无法获取用户信息

**可能原因**：
- LinuxDo API返回格式变化
- Access token无效

**调试**：
在 `internal/handlers/oauth.go` 中添加日志查看API响应

### 问题4：数据库迁移失败

**手动执行迁移**：
```bash
docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway -c "
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
"
```

---

## ✅ 验证清单

部署完成后，请验证以下项目：

- [ ] 环境变量已正确配置
- [ ] 数据库迁移已执行
- [ ] 服务已重启
- [ ] 登录页面显示LinuxDo登录按钮
- [ ] 点击按钮跳转到LinuxDo授权页面
- [ ] 授权后成功回调
- [ ] 用户信息正确显示
- [ ] 可以正常使用API

---

## 🎉 完成！

LinuxDo OAuth登录已成功集成！用户现在可以使用LinuxDo账户快速登录您的Codex Gateway。

### 优势
- ✅ 无需注册，一键登录
- ✅ 使用LinuxDo社区账户
- ✅ 自动同步用户名和头像
- ✅ 支持账户关联
- ✅ 安全的OAuth 2.0流程

### 下一步
- 考虑添加更多OAuth提供商（GitHub, Google等）
- 添加账户绑定/解绑功能
- 优化用户体验
