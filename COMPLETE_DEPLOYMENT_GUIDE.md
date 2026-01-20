# 🚀 完整部署指南 - LinuxDo OAuth + 定价修复

## 📋 本次更新内容

### 1. LinuxDo OAuth登录集成 ✅
- 支持使用LinuxDo账户一键登录
- 自动账户关联（通过邮箱）
- 同步用户名和头像

### 2. 定价修复 ✅
- 修正gpt-5.1-codex定价：$0.00125/$0.01 per 1K
- 修正gpt-5.2-codex定价：$0.00175/$0.014 per 1K
- 移除1.5倍markup，与Sub2API定价一致

### 3. 缓存Token显示 ✅
- 使用记录表格新增"缓存Token"列
- 缓存token显示为绿色高亮

### 4. Go版本修复 ✅
- 降级到Go 1.23.0以兼容Docker镜像

---

## 🔧 部署步骤

### 步骤1：SSH连接到服务器

```bash
ssh root@23.80.88.63
cd /root/codex-gateway
```

### 步骤2：更新环境变量

编辑 `.env` 文件：

```bash
nano .env
```

添加以下LinuxDo OAuth配置：

```env
# LinuxDo OAuth Configuration
LINUXDO_CLIENT_ID=kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ
LINUXDO_CLIENT_SECRET=XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg
LINUXDO_REDIRECT_URL=https://codex.zenscaleai.com/api/auth/linuxdo/callback
FRONTEND_URL=https://codex.zenscaleai.com
DEFAULT_BALANCE=0
```

保存并退出（Ctrl+X, Y, Enter）

### 步骤3：运行数据库迁移

```bash
docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway
```

在psql中执行：

```sql
-- 添加OAuth字段
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);

-- 密码字段改为可选
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- 创建OAuth索引
CREATE INDEX IF NOT EXISTS idx_oauth ON users(oauth_provider, oauth_id);

-- 更新现有用户
UPDATE users SET oauth_provider = 'email' WHERE oauth_provider IS NULL;

-- 退出
\q
```

### 步骤4：拉取最新代码并部署

```bash
git pull origin main
./deploy-auto.sh
```

等待部署完成（约2-3分钟）。

### 步骤5：验证部署

检查服务状态：

```bash
docker-compose ps
```

应该看到所有服务都是 `Up` 状态。

查看日志：

```bash
docker-compose logs -f backend | head -50
```

应该看到：
```
Server starting on port 12322
```

---

## 🧪 测试功能

### 1. 测试LinuxDo登录

1. 访问 https://codex.zenscaleai.com/login
2. 点击"使用LinuxDo登录"按钮
3. 在LinuxDo授权页面点击"授权"
4. 自动登录并跳转到Dashboard

### 2. 验证定价修复

发起一个测试请求：

```bash
curl -X POST https://codex.zenscaleai.com/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-5.2-codex",
    "messages": [{"role": "user", "content": "测试"}],
    "stream": true
  }'
```

然后查看使用记录，费用应该与Sub2API一致。

### 3. 验证缓存Token显示

访问 https://codex.zenscaleai.com/usage

应该看到表格中有"缓存Token"列。

---

## 📊 预期结果

### LinuxDo登录
- ✅ 登录页面显示LinuxDo登录按钮
- ✅ 点击后跳转到LinuxDo授权页面
- ✅ 授权后自动登录
- ✅ 用户名和头像正确显示

### 定价
- ✅ gpt-5.2-codex费用与Sub2API一致
- ✅ 无额外markup
- ✅ 缓存token正确计费（10%折扣）

### 使用记录
- ✅ 显示缓存Token列
- ✅ 缓存>0时显示为绿色
- ✅ 缓存=0时显示为灰色

---

## 🐛 故障排查

### 问题1：LinuxDo登录按钮无反应

**检查环境变量**：
```bash
docker-compose exec backend env | grep LINUXDO
```

应该看到：
```
LINUXDO_CLIENT_ID=kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ
LINUXDO_CLIENT_SECRET=XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg
...
```

如果没有，重新编辑 `.env` 并重启服务：
```bash
docker-compose restart backend
```

### 问题2：数据库迁移失败

**检查表结构**：
```bash
docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway -c "\d users"
```

应该看到 `oauth_provider`, `oauth_id`, `username`, `avatar_url` 字段。

如果没有，手动执行迁移SQL。

### 问题3：Docker构建失败

如果看到 "go.mod requires go >= 1.24.0" 错误：

```bash
# 确认已拉取最新代码
git log --oneline -1
```

应该看到：`15cd349 Fix: Downgrade Go version to 1.23.0 for Docker compatibility`

如果不是，执行：
```bash
git pull origin main
```

### 问题4：定价仍然不对

**检查数据库中的定价**：
```bash
docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway -c \
  "SELECT model_name, input_price_per_1k, output_price_per_1k, markup_multiplier
   FROM model_pricings
   WHERE model_name IN ('gpt-5.1-codex', 'gpt-5.2-codex');"
```

应该看到：
```
    model_name    | input_price_per_1k | output_price_per_1k | markup_multiplier
------------------+--------------------+---------------------+-------------------
 gpt-5.1-codex    |            0.00125 |                0.01 |                 1
 gpt-5.2-codex    |            0.00175 |               0.014 |                 1
```

如果不对，重启服务让seed函数重新执行：
```bash
docker-compose restart backend
```

---

## ✅ 验证清单

部署完成后，请逐项验证：

- [ ] 环境变量已配置（LINUXDO_*）
- [ ] 数据库迁移已执行（OAuth字段存在）
- [ ] 服务已重启并正常运行
- [ ] LinuxDo登录按钮显示
- [ ] LinuxDo登录流程正常
- [ ] 定价与Sub2API一致
- [ ] 缓存Token列显示
- [ ] 所有功能正常工作

---

## 🎉 完成！

恭喜！您的Codex Gateway现在支持：

1. ✅ **LinuxDo一键登录** - 无需注册，使用LinuxDo账户即可登录
2. ✅ **准确的定价** - 与Sub2API完全一致，无额外markup
3. ✅ **缓存Token显示** - 清晰展示缓存使用情况
4. ✅ **稳定的构建** - Go版本兼容Docker镜像

### 下一步建议

- 监控LinuxDo登录使用情况
- 考虑添加更多OAuth提供商（GitHub, Google等）
- 优化用户体验
- 添加账户绑定/解绑功能

---

**需要帮助？** 查看详细文档：
- LinuxDo OAuth: `LINUXDO_OAUTH_DEPLOYMENT.md`
- 定价修复: `DEPLOY_PRICING_FIX.md`
- 缓存Token: `CACHED_TOKENS_DISPLAY.md`
