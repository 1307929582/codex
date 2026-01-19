# 🎉 Codex Gateway - 真正的零配置部署

## 一键部署（3个命令）

```bash
git clone https://github.com/1307929582/codex.git
cd codex
./deploy-auto.sh
```

## 首次访问自动配置

部署完成后，访问 `http://localhost:12321`，系统会自动跳转到配置向导：

### 第1步：创建管理员账户
- 输入邮箱地址
- 设置密码（至少6个字符）
- 确认密码

### 第2步：配置OpenAI
- 输入OpenAI API密钥（从 [platform.openai.com/api-keys](https://platform.openai.com/api-keys) 获取）
- 可选：修改Base URL（如使用代理）

### 第3步：系统设置
- 可选：设置系统公告
- 可选：设置新用户默认余额
- 可选：是否允许新用户注册

点击"完成设置"后，自动登录到管理员面板！

---

## 完整流程演示

```bash
# 1. 克隆代码
git clone https://github.com/1307929582/codex.git
cd codex

# 2. 一键部署（自动生成所有配置）
./deploy-auto.sh

# 输出示例：
# ================================
#   Codex Gateway 一键部署
# ================================
#
# [1/7] 检查系统要求...
# ✓ Docker已安装
#
# [2/7] 停止现有服务...
#
# [3/7] 生成配置文件...
# ✓ 配置文件已生成
#
# [4/7] 构建Docker镜像...
#
# [5/7] 启动服务...
#
# [6/7] 等待服务就绪...
# 等待数据库启动 ✓
# 等待后端启动 ✓
#
# [7/7] 检查管理员账户...
# 未找到管理员账户
#
# ================================
# 部署完成！
# ================================
#
# 访问地址：
#   前端:        http://localhost:12321
#   管理员面板:  http://localhost:12321/admin
#   后端API:     http://localhost:12322

# 3. 打开浏览器
open http://localhost:12321

# 4. 自动跳转到 /setup 配置向导
# 5. 完成3步配置
# 6. 自动登录，开始使用！
```

---

## 零配置说明

### ❌ 不需要做的事情

- ❌ 不需要编辑 `.env` 文件
- ❌ 不需要手动生成密码
- ❌ 不需要执行SQL命令
- ❌ 不需要配置数据库
- ❌ 不需要重启服务

### ✅ 自动完成的事情

- ✅ 自动生成数据库密码（32字符随机）
- ✅ 自动生成JWT密钥（40字符随机）
- ✅ 自动创建数据库
- ✅ 自动执行数据库迁移
- ✅ 自动构建Docker镜像
- ✅ 自动启动所有服务
- ✅ 自动健康检查

### 🌐 Web界面完成的配置

- 🌐 创建管理员账户
- 🌐 配置OpenAI API密钥
- 🌐 配置系统参数
- 🌐 所有配置实时生效

---

## 技术细节

### 自动生成的配置

部署脚本会自动生成 `.env` 文件：

```bash
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<自动生成的32字符随机密码>
DB_NAME=codex_gateway
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=<自动生成的40字符随机密钥>

# Server Configuration
SERVER_PORT=8080

# Frontend Configuration
NEXT_PUBLIC_API_URL=http://localhost:12322
```

### 首次访问检测

系统会自动检测是否需要初始化：

1. 访问任何页面
2. 前端调用 `GET /api/setup/status`
3. 如果没有管理员账户，自动跳转到 `/setup`
4. 完成配置后，调用 `POST /api/setup/initialize`
5. 自动登录，跳转到管理员面板

---

## 常见问题

### Q: 如果我已经部署过，再次运行会怎样？

A: 脚本会检测现有配置：
- 如果 `.env` 存在，使用现有配置
- 如果已有管理员，跳过创建步骤
- 不会覆盖任何现有数据

### Q: 如何重置系统？

A: 删除数据并重新部署：

```bash
docker-compose down -v  # 删除所有数据
rm .env                 # 删除配置文件
./deploy-auto.sh        # 重新部署
```

### Q: 配置向导可以跳过吗？

A: 不可以。首次部署必须完成配置向导才能使用系统。这确保了：
- 至少有一个管理员账户
- OpenAI API已正确配置
- 系统参数已初始化

### Q: 配置完成后可以修改吗？

A: 可以！所有配置都可以在管理员面板修改：
- 访问 `/admin/settings`
- 修改OpenAI配置、系统参数等
- 实时生效，无需重启

---

## 对比其他部署方式

### 传统方式（需要10+步骤）

```bash
# 1. 克隆代码
git clone ...

# 2. 复制配置文件
cp .env.example .env

# 3. 编辑配置文件
nano .env
# 手动填写数据库密码
# 手动生成JWT密钥
# 手动配置OpenAI密钥

# 4. 启动服务
docker-compose up -d

# 5. 等待服务启动
sleep 30

# 6. 连接数据库
docker exec -it postgres psql ...

# 7. 创建管理员
UPDATE users SET role = 'admin' ...

# 8. 退出数据库
\q

# 9. 重新登录
# 10. 配置系统参数
```

### 现在的方式（3个命令）

```bash
git clone https://github.com/1307929582/codex.git
cd codex
./deploy-auto.sh
# 打开浏览器，完成3步配置向导
```

---

## 架构说明

### 部署架构

```
deploy-auto.sh
    ↓
自动生成 .env（数据库密码、JWT密钥）
    ↓
docker-compose build
    ↓
docker-compose up -d
    ↓
健康检查（数据库、后端）
    ↓
部署完成
```

### 首次访问流程

```
用户访问 http://localhost:12321
    ↓
前端检查 /api/setup/status
    ↓
没有管理员？→ 跳转到 /setup
    ↓
用户完成3步配置
    ↓
POST /api/setup/initialize
    ↓
创建管理员 + 保存OpenAI配置 + 保存系统设置
    ↓
返回JWT token
    ↓
自动登录 → 跳转到 /admin
```

---

## 总结

现在部署Codex Gateway只需要：

1. **运行一个脚本**：`./deploy-auto.sh`
2. **打开浏览器**：`http://localhost:12321`
3. **完成3步配置**：管理员 → OpenAI → 系统设置

就这么简单！

---

**最后更新**：2026-01-19
**版本**：v2.0.0 - WordPress风格的首次安装向导
