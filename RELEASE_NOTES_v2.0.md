# 🎉 Codex Gateway v2.0.0 发布说明

**发布日期**: 2026-01-19

这是Codex Gateway的重大版本更新，带来了革命性的零配置部署体验和完整的管理员面板。

---

## 🌟 重大更新

### 1. WordPress风格的首次安装向导

**全新的Web配置体验**！不再需要编辑配置文件或执行SQL命令。

#### 特性
- 🎯 **3步完成配置**：管理员账户 → OpenAI配置 → 系统设置
- 🚀 **自动跳转**：首次访问自动进入安装向导
- 🔐 **自动登录**：配置完成后自动登录管理员面板
- 💡 **友好界面**：进度条、表单验证、实时提示

#### 使用流程
```
访问 http://localhost:3000
    ↓
自动跳转到 /setup
    ↓
Step 1: 创建管理员账户
Step 2: 配置OpenAI API
Step 3: 系统设置（可选）
    ↓
自动登录 → 管理员面板
```

---

### 2. 零配置一键部署

**真正的一键部署**！只需3个命令即可完成整个系统的部署。

#### 部署命令
```bash
git clone https://github.com/1307929582/codex.git
cd codex
./deploy-auto.sh
```

#### 自动完成的任务
- ✅ 生成32字符随机数据库密码
- ✅ 生成40字符随机JWT密钥
- ✅ 创建并配置.env文件
- ✅ 构建Docker镜像
- ✅ 启动所有服务（数据库、后端、前端）
- ✅ 健康检查（等待服务就绪）
- ✅ 可选：交互式创建管理员账户

#### 无需手动操作
- ❌ 不需要编辑.env文件
- ❌ 不需要手动生成密码
- ❌ 不需要执行SQL命令
- ❌ 不需要配置数据库

---

### 3. 完整的管理员面板

**企业级管理功能**，提供全面的系统管理能力。

#### 3.1 Dashboard（系统概览）
- 📊 **统计卡片**：用户数、收入、消费、API密钥数
- 📈 **今日数据**：实时显示今日新增和变化
- 📢 **系统公告**：显示给所有用户

#### 3.2 用户管理
- 👥 **用户列表**：搜索、筛选、分页
- 👁 **用户详情**：基本信息、统计数据
- 💰 **余额调整**：充值/扣除，带说明
- ⏸ **状态管理**：暂停/激活用户

#### 3.3 系统设置
- 🔑 **OpenAI配置**：API密钥、Base URL
- 📢 **系统公告**：多行文本
- 💵 **用户设置**：默认余额、最小充值金额
- 🚪 **注册开关**：控制是否允许新用户注册

#### 3.4 操作日志
- 📝 **审计追踪**：所有管理员操作记录
- 🕐 **详细信息**：时间、操作、目标、详情、IP地址
- 📄 **分页浏览**：每页50条记录

---

### 4. OpenAI配置移至Web界面

**不再需要环境变量**！所有OpenAI配置都在管理员面板完成。

#### 配置位置
访问 `/admin/settings` → OpenAI配置

#### 配置项
- **OpenAI API密钥**：密码框，脱敏显示
- **OpenAI Base URL**：支持自定义代理

#### 优势
- ✅ **实时生效**：修改后立即生效，无需重启
- ✅ **更安全**：存储在数据库，不在环境变量
- ✅ **更灵活**：随时切换API密钥或代理

---

## 🔧 技术改进

### 后端改进

#### 新增API端点
- `GET /api/setup/status` - 检查是否需要初始化
- `POST /api/setup/initialize` - 完成首次设置
- `GET /api/admin/stats/overview` - 系统统计概览
- `GET /api/admin/users` - 用户列表
- `GET /api/admin/users/:id` - 用户详情
- `PUT /api/admin/users/:id/balance` - 调整余额
- `PUT /api/admin/users/:id/status` - 更新状态
- `GET /api/admin/settings` - 获取系统设置
- `PUT /api/admin/settings` - 更新系统设置
- `GET /api/admin/logs` - 操作日志

#### 数据模型更新
- `users` 表添加 `role` 字段（user/admin/super_admin）
- 新增 `system_settings` 表（系统配置）
- 新增 `admin_logs` 表（操作日志）
- `system_settings` 添加 `openai_api_key` 和 `openai_base_url` 字段

#### 配置简化
- 移除 `OPENAI_API_KEY` 环境变量
- 移除 `OPENAI_BASE_URL` 环境变量
- 从数据库读取OpenAI配置

### 前端改进

#### 新增页面
- `/setup` - 首次安装向导
- `/admin` - 管理员Dashboard
- `/admin/users` - 用户管理
- `/admin/users/[id]` - 用户详情
- `/admin/settings` - 系统设置
- `/admin/logs` - 操作日志

#### 新增组件
- `SetupRedirect` - 自动跳转到安装向导
- `AdminLayout` - 管理员布局和侧边栏
- 各种管理员页面组件

#### API客户端
- 新增 `adminApi` 模块（8个函数）
- 完整的TypeScript类型定义

---

## 📚 文档更新

### 新增文档
1. **README_DEPLOY.md** - 完整部署指南
2. **QUICK_START.md** - 快速开始指南
3. **DEPLOYMENT_FINAL.md** - 最终部署文档
4. **ADMIN_GUIDE.md** - 管理员使用指南
5. **FEATURES_DEMO.md** - 功能演示文档
6. **PROJECT_SUMMARY.md** - 项目总结
7. **DEPLOYMENT_CHECKLIST.md** - 部署检查清单

### 更新文档
- **README.md** - 更新为v2.0功能
- **.env.production.example** - 移除OpenAI配置

---

## 🔄 迁移指南

### 从v1.x升级到v2.0

#### 1. 备份数据
```bash
# 备份数据库
docker exec codex-postgres pg_dump -U postgres codex_gateway > backup.sql

# 备份环境变量
cp .env .env.backup
```

#### 2. 拉取最新代码
```bash
git pull origin main
```

#### 3. 更新环境变量
```bash
# 从.env中移除以下配置（已移至管理员面板）
# OPENAI_API_KEY=xxx
# OPENAI_BASE_URL=xxx
```

#### 4. 重新部署
```bash
docker-compose down
docker-compose build
docker-compose up -d
```

#### 5. 数据库迁移
后端启动时会自动执行迁移：
- 添加 `users.role` 字段
- 创建 `system_settings` 表
- 创建 `admin_logs` 表

#### 6. 配置OpenAI
1. 访问 `/admin/settings`
2. 在"OpenAI配置"中填写API密钥
3. 点击"保存设置"

#### 7. 提升管理员权限
```bash
docker exec -it codex-postgres psql -U postgres -d codex_gateway -c \
  "UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';"
```

---

## 🐛 Bug修复

### 安全修复（v1.0.1）
- 🔒 修复DoS攻击漏洞（请求体大小限制）
- 🔒 修复财务漏洞（原子化事务）
- 🔒 修复JWT安全问题（密钥长度验证）
- 🔒 添加请求限流保护

### 其他修复
- 修复用户注册时余额初始化问题
- 修复API密钥删除后仍可使用的问题
- 修复使用记录分页显示错误
- 优化数据库查询性能

---

## 📊 性能改进

### 后端性能
- 优化数据库查询（添加索引）
- 实现连接池管理
- 添加请求超时保护
- 优化事务处理

### 前端性能
- 使用TanStack Query缓存
- 实现乐观更新
- 优化组件渲染
- 代码分割和懒加载

---

## 🔐 安全增强

### 认证授权
- ✅ 角色权限控制（user/admin/super_admin）
- ✅ 管理员API双重验证（JWT + Role）
- ✅ 操作审计日志

### 数据保护
- ✅ 密码bcrypt加密
- ✅ API密钥哈希存储
- ✅ JWT密钥长度验证（至少32字符）

### 请求保护
- ✅ 请求体大小限制（1MB）
- ✅ 超时保护（60秒）
- ✅ 连接池限制

---

## 📈 统计数据

### 代码变更
- **新增文件**: 20+个
- **修改文件**: 15+个
- **新增代码**: ~4,000行
- **新增文档**: ~10,000字

### 功能统计
- **新增API端点**: 10个
- **新增页面**: 6个
- **新增组件**: 10+个
- **新增数据表**: 2个

---

## 🚀 未来计划

### v2.1（短期）
- [ ] 批量用户操作
- [ ] 数据导出功能
- [ ] 更详细的统计图表
- [ ] 邮件通知功能

### v2.2（中期）
- [ ] 多租户支持
- [ ] 自定义定价策略
- [ ] API速率限制配置
- [ ] Webhook支持

### v3.0（长期）
- [ ] 支持更多AI模型（Claude、Gemini等）
- [ ] 插件系统
- [ ] 白标解决方案
- [ ] 企业级SSO

---

## 🙏 致谢

感谢所有使用和支持Codex Gateway的用户！

特别感谢：
- 所有提交Issue和PR的贡献者
- 测试和反馈的早期用户
- 开源社区的支持

---

## 📞 获取帮助

### 文档
- [README.md](./README.md) - 项目主文档
- [README_DEPLOY.md](./README_DEPLOY.md) - 部署指南
- [ADMIN_GUIDE.md](./ADMIN_GUIDE.md) - 管理员指南
- [FEATURES_DEMO.md](./FEATURES_DEMO.md) - 功能演示

### 支持
- GitHub Issues: https://github.com/1307929582/codex/issues
- 文档站点: （待添加）

---

## 📄 许可证

MIT License

---

**发布版本**: v2.0.0
**发布日期**: 2026-01-19
**项目状态**: ✅ 生产就绪

---

## 🎉 立即开始

```bash
git clone https://github.com/1307929582/codex.git
cd codex
./deploy-auto.sh
```

访问 `http://localhost:3000` 开始使用！
