# 生产环境部署指南

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 2GB+ RAM
- 10GB+ 磁盘空间

## 快速部署

### 1. 克隆仓库

```bash
git clone https://github.com/1307929582/codex.git
cd codex
```

### 2. 使用部署CLI

```bash
./deploy/deploy.sh
```

选择选项 `1` 进行首次部署，脚本将自动：
- 检查系统要求
- 配置环境变量
- 构建Docker镜像
- 启动所有服务
- 初始化数据库

### 3. 访问服务

- **前端**: http://localhost:12321
- **后端API**: http://localhost:12322
- **健康检查**: http://localhost:12322/health

## 手动部署

### 1. 配置环境变量

```bash
cp .env.production.example .env
```

编辑 `.env` 文件，设置以下必需变量：

```env
DB_PASSWORD=your-secure-password
OPENAI_API_KEY=sk-your-openai-key
JWT_SECRET=your-jwt-secret-min-32-chars
NEXT_PUBLIC_API_URL=http://your-domain.com:12322
```

### 2. 构建并启动

```bash
docker-compose build
docker-compose up -d
```

### 3. 查看日志

```bash
docker-compose logs -f
```

### 4. 停止服务

```bash
docker-compose down
```

## 生产环境配置

### 使用Nginx反向代理

创建 `/etc/nginx/sites-available/codex-gateway`:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # Frontend
    location / {
        proxy_pass http://localhost:12321;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # Backend API
    location /api {
        proxy_pass http://localhost:12322;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

启用配置：

```bash
sudo ln -s /etc/nginx/sites-available/codex-gateway /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### SSL证书（Let's Encrypt）

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

## 数据备份

### 备份数据库

```bash
docker exec codex-postgres pg_dump -U postgres codex_gateway > backup.sql
```

### 恢复数据库

```bash
docker exec -i codex-postgres psql -U postgres codex_gateway < backup.sql
```

## 监控

### 查看服务状态

```bash
docker-compose ps
```

### 查看资源使用

```bash
docker stats
```

### 查看日志

```bash
# 所有服务
docker-compose logs -f

# 特定服务
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres
```

## 故障排查

### 后端无法连接数据库

检查数据库是否就绪：

```bash
docker-compose logs postgres
```

### 前端无法连接后端

检查 `NEXT_PUBLIC_API_URL` 环境变量是否正确。

### 端口冲突

修改 `docker-compose.yml` 中的端口映射：

```yaml
ports:
  - "8081:8080"  # 使用8081代替8080
```

## 更新部署

### 拉取最新代码

```bash
git pull origin main
```

### 重新构建并重启

```bash
docker-compose build
docker-compose up -d
```

## 安全建议

1. **修改默认密码**: 确保 `DB_PASSWORD` 和 `JWT_SECRET` 使用强密码
2. **启用HTTPS**: 生产环境必须使用SSL证书
3. **防火墙配置**: 只开放必要的端口（80, 443）
4. **定期备份**: 设置自动备份任务
5. **监控日志**: 定期检查异常日志
6. **更新依赖**: 定期更新Docker镜像和依赖包

## 性能优化

### 数据库连接池

编辑 `internal/database/database.go`，配置连接池：

```go
sqlDB, _ := DB.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### 前端CDN

将静态资源部署到CDN以提高加载速度。

## 扩展部署

### 多实例部署

使用Docker Swarm或Kubernetes进行水平扩展。

### 负载均衡

使用Nginx或HAProxy进行负载均衡。

## 支持

如有问题，请查看：
- [GitHub Issues](https://github.com/1307929582/codex/issues)
- [API文档](./API_DOCUMENTATION.md)
- [项目README](./README.md)
