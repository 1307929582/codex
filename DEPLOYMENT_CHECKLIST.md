# ✅ Codex Gateway 部署检查清单

在部署Codex Gateway之前，请确保满足以下要求。

---

## 📋 系统要求

### 必需软件

- [ ] **Docker** (版本 20.10+)
  ```bash
  docker --version
  # 应输出: Docker version 20.10.x 或更高
  ```

- [ ] **Docker Compose** (版本 2.0+)
  ```bash
  docker-compose --version
  # 应输出: Docker Compose version 2.x.x 或更高
  ```

- [ ] **Git**
  ```bash
  git --version
  # 应输出: git version 2.x.x
  ```

### 系统资源

- [ ] **CPU**: 至少2核
- [ ] **内存**: 至少4GB RAM
- [ ] **磁盘**: 至少10GB可用空间
- [ ] **网络**: 能够访问GitHub和Docker Hub

### 端口要求

- [ ] **3000端口**: 前端服务（可修改）
- [ ] **8080端口**: 后端API（可修改）
- [ ] **5432端口**: PostgreSQL（内部使用）

检查端口是否被占用：
```bash
# Linux/Mac
lsof -i :3000
lsof -i :8080

# Windows
netstat -ano | findstr :3000
netstat -ano | findstr :8080
```

---

## 🚀 部署步骤检查清单

### 第一步：克隆代码

- [ ] 克隆仓库
  ```bash
  git clone https://github.com/1307929582/codex.git
  cd codex
  ```

- [ ] 确认文件完整
  ```bash
  ls -la
  # 应该看到: deploy-auto.sh, docker-compose.yml, README.md 等
  ```

### 第二步：运行部署脚本

- [ ] 添加执行权限
  ```bash
  chmod +x deploy-auto.sh
  ```

- [ ] 运行部署脚本
  ```bash
  ./deploy-auto.sh
  ```

- [ ] 等待部署完成（约2-5分钟）

### 第三步：验证服务

- [ ] 检查容器状态
  ```bash
  docker-compose ps
  # 所有容器应该是 "Up" 状态
  ```

- [ ] 检查后端健康
  ```bash
  curl http://localhost:8080/health
  # 应返回: {"status":"ok"}
  ```

- [ ] 检查前端访问
  ```bash
  curl -I http://localhost:3000
  # 应返回: HTTP/1.1 200 OK
  ```

### 第四步：完成安装向导

- [ ] 打开浏览器访问 `http://localhost:3000`
- [ ] 自动跳转到 `/setup` 页面
- [ ] 完成3步配置：
  - [ ] Step 1: 创建管理员账户
  - [ ] Step 2: 配置OpenAI API密钥
  - [ ] Step 3: 系统设置（可选）
- [ ] 点击"完成设置"
- [ ] 自动登录到管理员面板

### 第五步：验证功能

- [ ] 访问管理员面板 `http://localhost:3000/admin`
- [ ] 查看Dashboard统计数据
- [ ] 访问系统设置，确认OpenAI配置已保存
- [ ] 退出登录，重新登录测试

---

## 🔍 故障排查检查清单

### 问题1：部署脚本失败

- [ ] 检查Docker是否运行
  ```bash
  docker info
  ```

- [ ] 检查Docker Compose是否安装
  ```bash
  docker-compose --version
  ```

- [ ] 检查磁盘空间
  ```bash
  df -h
  ```

- [ ] 查看错误日志
  ```bash
  cat deploy-auto.log  # 如果存在
  ```

### 问题2：容器无法启动

- [ ] 查看容器日志
  ```bash
  docker-compose logs backend
  docker-compose logs frontend
  docker-compose logs postgres
  ```

- [ ] 检查端口占用
  ```bash
  lsof -i :3000
  lsof -i :8080
  lsof -i :5432
  ```

- [ ] 重新构建镜像
  ```bash
  docker-compose down
  docker-compose build --no-cache
  docker-compose up -d
  ```

### 问题3：数据库连接失败

- [ ] 检查PostgreSQL容器状态
  ```bash
  docker-compose ps postgres
  ```

- [ ] 检查数据库日志
  ```bash
  docker-compose logs postgres
  ```

- [ ] 测试数据库连接
  ```bash
  docker exec -it codex-postgres psql -U postgres -d codex_gateway -c "SELECT 1;"
  ```

### 问题4：前端无法访问

- [ ] 检查前端容器日志
  ```bash
  docker-compose logs frontend
  ```

- [ ] 检查环境变量
  ```bash
  docker exec codex-frontend env | grep NEXT_PUBLIC
  ```

- [ ] 重启前端容器
  ```bash
  docker-compose restart frontend
  ```

### 问题5：后端API错误

- [ ] 检查后端日志
  ```bash
  docker-compose logs backend
  ```

- [ ] 检查环境变量
  ```bash
  docker exec codex-backend env | grep -E "DB_|JWT_"
  ```

- [ ] 测试API健康检查
  ```bash
  curl http://localhost:8080/health
  ```

### 问题6：无法访问安装向导

- [ ] 清除浏览器缓存
- [ ] 使用无痕模式访问
- [ ] 检查前端日志
  ```bash
  docker-compose logs frontend
  ```

- [ ] 手动访问 `http://localhost:3000/setup`

---

## 🔐 安全检查清单

### 部署前

- [ ] 确认服务器防火墙配置
- [ ] 确认只开放必要端口（3000, 8080）
- [ ] 准备好OpenAI API密钥

### 部署后

- [ ] 修改默认数据库密码（如果需要）
- [ ] 确认JWT密钥已自动生成（40字符）
- [ ] 在安装向导中设置强密码
- [ ] 配置OpenAI API密钥
- [ ] 测试管理员登录
- [ ] 检查操作日志功能

### 生产环境额外检查

- [ ] 配置HTTPS（使用Nginx + Let's Encrypt）
- [ ] 设置数据库备份
- [ ] 配置日志轮转
- [ ] 设置监控告警
- [ ] 限制管理员面板访问IP（可选）

---

## 📊 部署后验证清单

### 功能验证

- [ ] **用户注册**
  - 访问 `http://localhost:3000/register`
  - 注册新用户
  - 确认能够登录

- [ ] **API密钥管理**
  - 创建新API密钥
  - 复制密钥
  - 测试API调用

- [ ] **OpenAI代理**
  ```bash
  curl -X POST http://localhost:8080/v1/chat/completions \
    -H "Authorization: Bearer YOUR_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}'
  ```

- [ ] **计费系统**
  - 检查余额是否正确扣除
  - 查看使用记录
  - 确认费用计算正确

- [ ] **管理员功能**
  - 查看用户列表
  - 调整用户余额
  - 修改系统设置
  - 查看操作日志

### 性能验证

- [ ] 测试并发请求
  ```bash
  # 使用ab或wrk进行压力测试
  ab -n 100 -c 10 http://localhost:8080/health
  ```

- [ ] 检查响应时间
- [ ] 监控内存使用
  ```bash
  docker stats
  ```

---

## 📝 部署记录

### 部署信息

- **部署日期**: _______________
- **部署人员**: _______________
- **服务器IP**: _______________
- **域名**: _______________

### 配置信息

- **前端URL**: http://localhost:3000
- **后端URL**: http://localhost:8080
- **数据库**: PostgreSQL 15
- **管理员邮箱**: _______________

### 检查结果

- [ ] 所有容器正常运行
- [ ] 安装向导完成
- [ ] 管理员账户创建成功
- [ ] OpenAI配置完成
- [ ] API调用测试通过
- [ ] 计费系统正常
- [ ] 所有功能验证通过

---

## 🎉 部署完成

恭喜！如果所有检查项都已完成，您的Codex Gateway已成功部署！

### 下一步

1. **阅读文档**
   - [ADMIN_GUIDE.md](./ADMIN_GUIDE.md) - 管理员使用指南
   - [FEATURES_DEMO.md](./FEATURES_DEMO.md) - 功能演示
   - [API_DOCUMENTATION.md](./API_DOCUMENTATION.md) - API文档

2. **配置生产环境**（如果需要）
   - 配置HTTPS
   - 设置域名
   - 配置备份
   - 设置监控

3. **开始使用**
   - 创建用户账户
   - 生成API密钥
   - 开始调用OpenAI API

---

## 📞 获取帮助

如果遇到问题：

1. 查看 [README_DEPLOY.md](./README_DEPLOY.md) 详细部署指南
2. 查看 [FEATURES_DEMO.md](./FEATURES_DEMO.md) 功能演示
3. 提交GitHub Issue: https://github.com/1307929582/codex/issues

---

**检查清单版本**: v2.0.0
**最后更新**: 2026-01-19
