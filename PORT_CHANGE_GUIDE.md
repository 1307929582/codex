# 🔄 端口配置变更说明

**变更日期**: 2026-01-19
**版本**: v2.0.1

---

## 📊 端口变更

### 旧端口配置
```
前端:     3000
后端API:  8080
数据库:   5432 (内部，未变更)
```

### 新端口配置
```
前端:     12321
后端API:  12322
数据库:   5432 (内部，未变更)
```

---

## 🔧 已更新的文件

### 配置文件（3个）
1. ✅ `docker-compose.yml` - Docker服务端口映射
2. ✅ `.env.production.example` - 环境变量示例
3. ✅ `deploy-auto.sh` - 部署脚本

### 文档文件（21个）
1. ✅ README.md
2. ✅ QUICK_START.md
3. ✅ README_DEPLOY.md
4. ✅ ADMIN_GUIDE.md
5. ✅ FEATURES_DEMO.md
6. ✅ QUICK_REFERENCE.md
7. ✅ API_DOCUMENTATION.md
8. ✅ PROJECT_SUMMARY.md
9. ✅ RELEASE_NOTES_v2.0.md
10. ✅ DEPLOYMENT_CHECKLIST.md
11. ✅ DEPLOYMENT_FINAL.md
12. ✅ DEPLOYMENT.md
13. ✅ DELIVERY_CHECKLIST.md
14. ✅ DELIVERY_REPORT.md
15. ✅ PROJECT_FINAL_SUMMARY.md
16. ✅ PROJECT_SHOWCASE.md
17. ✅ SECURITY_FIXES.md
18. ✅ frontend/README.md
19. ✅ 项目完成报告.md
20. ✅ 其他相关文档

---

## 🚀 如何应用更新

### 对于新部署

直接使用新的部署命令：

```bash
git clone https://github.com/1307929582/codex.git
cd codex
./deploy-auto.sh
```

访问地址：
- 前端: http://localhost:12321
- 管理员面板: http://localhost:12321/admin
- 后端API: http://localhost:12322

### 对于现有部署

如果您已经部署了旧版本，需要执行以下步骤：

#### 1. 拉取最新代码
```bash
cd /path/to/codex
git pull origin main
```

#### 2. 停止现有服务
```bash
docker-compose down
```

#### 3. 更新环境变量（如果有自定义.env）
如果您有自定义的 `.env` 文件，需要更新：

```bash
# 旧配置
NEXT_PUBLIC_API_URL=http://localhost:8080

# 新配置
NEXT_PUBLIC_API_URL=http://localhost:12322
```

#### 4. 重新构建并启动
```bash
docker-compose up -d --build
```

#### 5. 验证服务
```bash
# 检查容器状态
docker-compose ps

# 测试前端
curl -I http://localhost:12321

# 测试后端
curl http://localhost:12322/health
```

---

## 🔍 验证清单

### 服务可访问性
- [ ] 前端可访问: http://localhost:12321
- [ ] 管理员面板可访问: http://localhost:12321/admin
- [ ] 后端API可访问: http://localhost:12322
- [ ] 健康检查通过: http://localhost:12322/health

### 功能验证
- [ ] 用户可以正常登录
- [ ] 管理员面板正常工作
- [ ] API调用正常
- [ ] OpenAI代理正常

---

## ⚠️ 注意事项

### 1. 端口冲突
如果新端口（12321或12322）被占用，请检查：

```bash
# 检查端口占用（Mac/Linux）
lsof -i :12321
lsof -i :12322

# 检查端口占用（Windows）
netstat -ano | findstr :12321
netstat -ano | findstr :12322
```

### 2. 防火墙配置
如果部署在服务器上，确保防火墙允许新端口：

```bash
# Ubuntu/Debian
sudo ufw allow 12321
sudo ufw allow 12322

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=12321/tcp
sudo firewall-cmd --permanent --add-port=12322/tcp
sudo firewall-cmd --reload
```

### 3. Nginx反向代理
如果使用Nginx，需要更新配置：

```nginx
# 旧配置
proxy_pass http://localhost:3000;  # 前端
proxy_pass http://localhost:8080;  # 后端

# 新配置
proxy_pass http://localhost:12321;  # 前端
proxy_pass http://localhost:12322;  # 后端
```

### 4. API客户端
如果有外部客户端调用API，需要更新API地址：

```javascript
// 旧配置
const API_URL = 'http://localhost:8080';

// 新配置
const API_URL = 'http://localhost:12322';
```

---

## 🐛 故障排查

### 问题1: 容器无法启动

**症状**: docker-compose up 失败

**解决方案**:
```bash
# 清理旧容器
docker-compose down -v

# 重新构建
docker-compose build --no-cache

# 启动
docker-compose up -d
```

### 问题2: 前端无法连接后端

**症状**: 前端显示网络错误

**解决方案**:
```bash
# 检查环境变量
docker exec codex-frontend env | grep NEXT_PUBLIC

# 应该显示:
# NEXT_PUBLIC_API_URL=http://localhost:12322

# 如果不正确，重新构建前端
docker-compose build frontend
docker-compose up -d frontend
```

### 问题3: 旧端口仍在使用

**症状**: 服务仍在旧端口运行

**解决方案**:
```bash
# 确保旧容器已停止
docker ps -a | grep codex

# 删除旧容器
docker rm -f codex-frontend codex-backend codex-postgres

# 重新部署
docker-compose up -d
```

---

## 📝 变更原因

用户要求将端口修改为：
- 前端: 12321
- 后端: 12322

这样可以避免与其他服务的端口冲突。

---

## 📞 获取帮助

如果遇到问题：

1. 查看日志
   ```bash
   docker-compose logs -f
   ```

2. 检查文档
   - [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - 常用命令
   - [DEPLOYMENT_CHECKLIST.md](./DEPLOYMENT_CHECKLIST.md) - 部署检查清单

3. 提交Issue
   - https://github.com/1307929582/codex/issues

---

## ✅ 变更确认

- ✅ 所有配置文件已更新
- ✅ 所有文档已更新
- ✅ 部署脚本已更新
- ✅ 代码已提交到GitHub

**变更已完成，可以正常使用！**

---

**更新日期**: 2026-01-19
**Git提交**: b905e31
**状态**: ✅ 完成
