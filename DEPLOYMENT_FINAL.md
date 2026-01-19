# 🚀 最终部署指南（无环境变量配置）

## ✅ 最小化环境变量

现在只需要配置**3个**环境变量即可启动系统！

### 必需的环境变量

```bash
# .env 文件
DB_PASSWORD=your-secure-password
JWT_SECRET=your-jwt-secret-min-32-chars
NEXT_PUBLIC_API_URL=http://localhost:12322
```

### ❌ 不再需要的环境变量

```bash
# 以下配置已移到管理员面板
OPENAI_API_KEY=xxx  # ❌ 删除
OPENAI_BASE_URL=xxx # ❌ 删除
```

---

## 📦 完整部署步骤

### 1. 拉取最新代码

```bash
cd /path/to/codex中转
git pull origin main
```

### 2. 配置环境变量

```bash
# 复制示例文件
cp .env.production.example .env

# 编辑.env文件，只需配置3个变量
nano .env
```

```.env
DB_PASSWORD=your-secure-password
JWT_SECRET=$(openssl rand -base64 32)
NEXT_PUBLIC_API_URL=http://localhost:12322
```

### 3. 停止并重新构建

```bash
# 停止现有服务
docker-compose down

# 重新构建镜像
docker-compose build

# 启动所有服务
docker-compose up -d
```

### 4. 等待服务就绪

```bash
# 查看服务状态
docker-compose ps

# 查看后端日志（确认数据库迁移成功）
docker-compose logs -f backend
```

### 5. 创建管理员账户

**方式一：注册后提升**

```bash
# 1. 访问 http://localhost:12321 注册账户
# 2. 提升为管理员
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c "UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';"
```

**方式二：直接在数据库创建**

```bash
docker exec -it codex-postgres psql -U postgres -d codex_gateway

-- 查看现有用户
SELECT id, email, role FROM users;

-- 提升为管理员
UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';

-- 或创建超级管理员
UPDATE users SET role = 'super_admin' WHERE email = 'your-email@example.com';

-- 退出
\q
```

### 6. 配置OpenAI（管理员面板）

1. 访问 `http://localhost:12321`
2. 使用管理员账户登录
3. 访问 `http://localhost:12321/admin/settings`
4. 在"OpenAI配置"部分填写：
   - **OpenAI API密钥**：`sk-your-key-here`
   - **OpenAI Base URL**：`https://api.openai.com/v1`（或自定义代理）
5. 点击"保存设置"

### 7. 验证部署

```bash
# 测试管理员API
curl http://localhost:12322/api/admin/stats/overview \
  -H "Authorization: Bearer YOUR_TOKEN"

# 测试OpenAI代理（需要先配置OpenAI密钥）
curl -X POST http://localhost:12322/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}'
```

---

## 🎯 配置对比

### 之前（需要环境变量）

```bash
# 必须在.env文件配置
DB_PASSWORD=xxx
OPENAI_API_KEY=sk-xxx  # ❌ 每次修改需要重启
OPENAI_BASE_URL=xxx    # ❌ 每次修改需要重启
JWT_SECRET=xxx
```

### 现在（管理员面板配置）

```bash
# .env文件（只需3个）
DB_PASSWORD=xxx
JWT_SECRET=xxx
NEXT_PUBLIC_API_URL=xxx

# 管理员面板配置（无需重启）
- OpenAI API Key ✅ 实时生效
- OpenAI Base URL ✅ 实时生效
- 系统公告 ✅ 实时生效
- 新用户默认余额 ✅ 实时生效
- 最小充值金额 ✅ 实时生效
- 注册开关 ✅ 实时生效
```

---

## 💡 优势

### 1. 无需重启服务
- 修改OpenAI密钥：直接在管理员面板更新，立即生效
- 切换OpenAI代理：直接修改Base URL，无需重启
- 调整系统参数：所有配置实时生效

### 2. 更安全
- OpenAI密钥存储在数据库（加密）
- 不会出现在环境变量或日志中
- 只有管理员可以查看和修改

### 3. 更灵活
- 支持多个OpenAI账户切换
- 支持自定义代理
- 支持动态调整系统参数

### 4. 更简单
- 部署时只需配置3个环境变量
- 其他配置通过Web界面完成
- 无需SSH到服务器修改配置

---

## 🔧 常见问题

### Q1: 首次启动时OpenAI API调用失败？

**A**: 这是正常的！首次启动时数据库中没有OpenAI配置。

**解决方案**：
1. 登录管理员面板
2. 访问 `/admin/settings`
3. 配置OpenAI API密钥
4. 保存后即可正常使用

### Q2: 如何切换OpenAI账户？

**A**: 直接在管理员面板修改：
1. 访问 `/admin/settings`
2. 更新"OpenAI API密钥"
3. 点击保存
4. 立即生效，无需重启

### Q3: 如何使用自定义OpenAI代理？

**A**: 修改Base URL：
1. 访问 `/admin/settings`
2. 将"OpenAI Base URL"改为代理地址
3. 例如：`https://your-proxy.com/v1`
4. 保存后立即生效

### Q4: 数据库迁移失败？

**A**: 手动执行迁移：

```sql
-- 连接数据库
docker exec -it codex-postgres psql -U postgres -d codex_gateway

-- 添加OpenAI配置字段
ALTER TABLE system_settings
ADD COLUMN IF NOT EXISTS openai_api_key VARCHAR(255),
ADD COLUMN IF NOT EXISTS openai_base_url VARCHAR(255) DEFAULT 'https://api.openai.com/v1';
```

### Q5: 如何备份OpenAI配置？

**A**: 备份数据库即可：

```bash
# 备份整个数据库
docker exec codex-postgres pg_dump -U postgres codex_gateway > backup.sql

# 只备份系统设置
docker exec codex-postgres psql -U postgres -d codex_gateway -c "SELECT * FROM system_settings;" > settings_backup.txt
```

---

## 📊 配置清单

### 环境变量（.env文件）

| 变量 | 必需 | 说明 |
|------|------|------|
| `DB_PASSWORD` | ✅ | 数据库密码 |
| `JWT_SECRET` | ✅ | JWT签名密钥（至少32字符） |
| `NEXT_PUBLIC_API_URL` | ✅ | 前端API地址 |

### 管理员面板配置（/admin/settings）

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| OpenAI API密钥 | OpenAI API Key | 空（必须配置） |
| OpenAI Base URL | API端点地址 | https://api.openai.com/v1 |
| 系统公告 | 显示在Dashboard | 空 |
| 新用户默认余额 | 注册时赠送 | 0 |
| 最小充值金额 | 充值限制 | 10 |
| 注册开关 | 是否允许注册 | 开启 |

---

## 🎉 总结

现在您只需要：

1. **配置3个环境变量**（数据库、JWT、前端URL）
2. **启动服务**（docker-compose up -d）
3. **创建管理员**（一条SQL命令）
4. **在Web界面配置OpenAI**（无需重启）

所有业务配置都在管理员面板完成，无需再修改环境变量或重启服务！

---

**最后更新**：2026-01-19
**版本**：v1.2.0
