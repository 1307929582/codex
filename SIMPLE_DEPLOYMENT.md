# 🚀 一键部署指南 - 无需配置.env！

## ✨ 重大改进

**不再需要手动编辑.env文件！** 所有配置都通过Admin Panel管理。

---

## 📋 本次更新

1. ✅ **LinuxDo OAuth登录** - 通过Admin Panel配置
2. ✅ **定价修复** - 与Sub2API完全一致
3. ✅ **缓存Token显示** - 使用记录中显示缓存情况
4. ✅ **Go版本修复** - Docker构建兼容性

---

## 🚀 部署步骤（超简单！）

### 步骤1：SSH连接

```bash
ssh root@23.80.88.63
cd /root/codex-gateway
```

### 步骤2：运行数据库迁移

```bash
# 连接数据库
docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway

# 执行迁移（复制粘贴以下所有内容）
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
CREATE INDEX IF NOT EXISTS idx_oauth ON users(oauth_provider, oauth_id);
UPDATE users SET oauth_provider = 'email' WHERE oauth_provider IS NULL;

ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_id VARCHAR(255);
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_secret VARCHAR(255);
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_enabled BOOLEAN DEFAULT false;

UPDATE system_settings
SET
    linuxdo_client_id = 'kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ',
    linuxdo_client_secret = 'XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg',
    linuxdo_enabled = true
WHERE id = 1;

\q
```

### 步骤3：部署

```bash
git pull origin main
./deploy-auto.sh
```

**就这么简单！** 🎉

---

## ✅ 验证部署

### 1. 检查服务状态

```bash
docker-compose ps
```

所有服务应该都是 `Up` 状态。

### 2. 测试LinuxDo登录

1. 访问 https://codex.zenscaleai.com/login
2. 点击"使用LinuxDo登录"按钮
3. 授权后自动登录

### 3. 查看Admin Panel配置

1. 访问 https://codex.zenscaleai.com/admin/settings
2. 滚动到"LinuxDo OAuth"部分
3. 确认配置已自动填充

---

## 🎛️ 通过Admin Panel管理

### LinuxDo OAuth配置

访问：https://codex.zenscaleai.com/admin/settings

在"LinuxDo OAuth"部分，您可以：

- ✅ **启用/禁用** LinuxDo登录（一键开关）
- ✅ **修改Client ID** 和 **Client Secret**
- ✅ **查看回调地址** 配置说明

**无需重启服务，配置立即生效！**

### 其他配置

同样在Admin Panel中管理：

- **用户入门**：初始余额、最小充值、注册开关
- **Codex上游**：多上游管理、健康检查
- **系统公告**：全局通知

---

## 🐛 故障排查

### 问题：LinuxDo登录按钮无反应

**检查Admin Panel配置**：

1. 访问 https://codex.zenscaleai.com/admin/settings
2. 确认"启用LinuxDo登录"开关已打开
3. 确认Client ID和Secret已填写

### 问题：定价仍然不对

**重启服务让定价更新**：

```bash
docker-compose restart backend
```

### 问题：数据库迁移失败

**检查表结构**：

```bash
docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway -c "\d system_settings"
```

应该看到 `linuxdo_client_id`, `linuxdo_client_secret`, `linuxdo_enabled` 字段。

---

## 🎉 完成！

现在您的Codex Gateway：

1. ✅ **无需.env配置** - 所有配置通过Admin Panel管理
2. ✅ **LinuxDo一键登录** - 自动配置，开箱即用
3. ✅ **准确定价** - 与Sub2API完全一致
4. ✅ **缓存Token显示** - 清晰展示缓存使用

### 管理配置

所有配置都在Admin Panel：

- 🔗 https://codex.zenscaleai.com/admin/settings

**再也不用编辑.env文件了！** 🎊
