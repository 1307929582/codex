# 🚀 超简单部署 - 两条命令搞定！

## ✨ 完全自动化

**不需要任何手动配置！**

- ❌ 不需要编辑 .env
- ❌ 不需要运行 SQL
- ❌ 不需要配置 OAuth
- ✅ 只需要两条命令！

---

## 🎯 部署步骤

```bash
ssh root@23.80.88.63
cd /root/codex-gateway
git pull && ./deploy-auto.sh
```

**就这么简单！** 🎉

---

## 🔄 自动完成的事情

部署脚本会自动：

1. ✅ 拉取最新代码
2. ✅ 构建Docker镜像
3. ✅ 启动服务
4. ✅ 运行数据库迁移
5. ✅ 添加OAuth字段
6. ✅ 配置LinuxDo OAuth（自动填充Client ID和Secret）
7. ✅ 启用LinuxDo登录

**所有配置都自动完成！**

---

## ✅ 验证部署

### 1. 检查服务状态

```bash
docker-compose ps
```

所有服务应该都是 `Up` 状态。

### 2. 查看迁移日志

```bash
docker-compose logs backend | grep -i migration
```

应该看到：
```
Running database migrations...
Running migration 002: Add OAuth fields to users table
Migration 002: Completed successfully
Running migration 003: Add LinuxDo OAuth settings
Migration 003: LinuxDo OAuth auto-configured with default credentials
Migration 003: Completed successfully
All migrations completed successfully
```

### 3. 测试LinuxDo登录

1. 访问 https://codex.zenscaleai.com/login
2. 点击"使用LinuxDo登录"按钮
3. 授权后自动登录

### 4. 查看Admin Panel

访问 https://codex.zenscaleai.com/admin/settings

在"LinuxDo OAuth"部分，您会看到：
- ✅ LinuxDo登录已启用
- ✅ Client ID已自动填充
- ✅ Client Secret已自动填充

---

## 🎛️ 后续管理

### 修改LinuxDo OAuth配置

访问：https://codex.zenscaleai.com/admin/settings

在"LinuxDo OAuth"部分，您可以：
- 启用/禁用LinuxDo登录
- 修改Client ID和Secret
- 查看回调地址配置

**无需重启服务，配置立即生效！**

---

## 🐛 故障排查

### 问题：服务启动失败

**查看日志**：
```bash
docker-compose logs backend
```

### 问题：迁移失败

**重启服务重新运行迁移**：
```bash
docker-compose restart backend
```

迁移是幂等的，可以安全地多次运行。

### 问题：LinuxDo���录不工作

**检查Admin Panel配置**：
1. 访问 https://codex.zenscaleai.com/admin/settings
2. 确认"启用LinuxDo登录"开关已打开
3. 确认Client ID和Secret已填写

---

## 📋 本次更新内容

1. ✅ **LinuxDo OAuth登录** - 自动配置，开箱即用
2. ✅ **定价修复** - 与Sub2API完全一致
3. ✅ **缓存Token显示** - 使用记录中显示
4. ✅ **自动数据库迁移** - 无需手动执行SQL
5. ✅ **Admin Panel配置** - 所有配置通过UI管理

---

## 🎉 完成！

现在您的Codex Gateway：

- ✅ **零配置部署** - `git pull && ./deploy-auto.sh`
- ✅ **自动迁移** - 数据库自动更新
- ✅ **LinuxDo登录** - 自动配置，立即可用
- ✅ **Admin Panel管理** - 所有配置通过UI

**真正的一键部署！** 🚀

---

## 📝 下次更新

下次更新同样简单：

```bash
ssh root@23.80.88.63
cd /root/codex-gateway
git pull && ./deploy-auto.sh
```

就这么简单！
