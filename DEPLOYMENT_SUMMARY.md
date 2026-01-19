# Codex Gateway 部署总结

## 已完成的工作

### 1. 核心功能实现

#### ✅ 多端点支持
- `/v1/chat/completions` - ChatGPT API
- `/v1/responses` - Codex/Responses API
- `/v1/completions` - Legacy Codex API
- `/v1/engines/:engine/completions` - Engine-specific API
- `/v1/edits` - Edits API
- `/v1/embeddings` - Embeddings API

#### ✅ 智能路由
- ChatGPT 端点：应用转换（添加 instructions、stream_options）
- Codex 端点：原样转发（不做任何修改）
- 保留原始请求路径转发到上游

#### ✅ 上游管理
- 多上游支持（可配置多个 Sub2API 实例）
- 用户会话亲和性（一致性哈希）
- 健康检查（自动检测上游可用性）
- 自动故障转移

#### ✅ 计费系统
- 基于 token 使用量计费
- 支持流式和非流式请求
- 自动 token 估算（当 usage 信息缺失时）
- 完整的使用日志

### 2. 代码修改清单

#### 后端修改

**文件**: `cmd/gateway/main.go`
- 添加了 6 个新的 API 端点路由
- 支持 Codex、Edits、Embeddings API

**文件**: `internal/handlers/proxy.go`
- 修改 `ProxyHandler` 以区分 ChatGPT 和 Codex 端点
- 只对 `/chat/completions` 应用转换
- 添加 `requestPath` 参数传递
- 修改 `handleStreamingRequest` 和 `handleNonStreamingRequest`
- 只对 ChatGPT API 添加 `stream_options`

**文件**: `internal/handlers/admin_health.go`
- 改进健康检查反馈（返回检查的上游数量）

**文件**: `internal/upstream/health_checker.go`
- 添加详细的调试日志
- 记录每个上游的检查状态
- 显示请求 URL 和 API Key 前缀

#### 前端修改

**文件**: `frontend/src/app/admin/upstreams/page.tsx`
- 添加健康检查视觉反馈
- 显示检查状态消息
- 改进用户体验

**文件**: `frontend/src/app/admin/settings/page.tsx`
- 移除旧的 Codex 配置表单
- 添加迁移提示框
- 引导用户到新的上游管理页面

**文件**: `frontend/src/app/admin/layout.tsx`
- 添加"Codex 上游"导航链接

### 3. 文档

- ✅ `CODEX_SETUP_GUIDE.md` - 完整配置指南
- ✅ `DEPLOYMENT_SUMMARY.md` - 本文档

## 部署步骤

### 步骤 1: 更新服务器代码

```bash
# SSH 登录服务器
ssh root@23.80.88.63

# 进入项目目录
cd /root/codex-gateway

# 拉取最新代码
git pull origin main

# 查看更新内容
git log --oneline -10

# 应该看到以下提交：
# - Add comprehensive Codex setup guide
# - Fix Codex API support based on sub2api analysis
# - Improve health check feedback with detailed status
# - Add visual feedback for health check trigger
# - Remove deprecated Codex config from system settings
# - Add support for Codex API endpoints
# - Add detailed logging for health check debugging
```

### 步骤 2: 部署更新

```bash
# 运行部署脚本
./deploy-auto.sh

# 等待部署完成（约 1-2 分钟）
# 应该看到：
# - Building backend...
# - Building frontend...
# - Restarting services...
# - Deployment completed successfully
```

### 步骤 3: 验证服务状态

```bash
# 检查服务是否运行
docker-compose ps

# 应该看到所有服务都是 Up 状态：
# - codex-gateway-backend-1
# - codex-gateway-frontend-1
# - codex-gateway-db-1

# 查看后端日志
docker-compose logs -f backend | head -50

# 应该看到：
# - Server starting on port 8080
# - [HealthCheck] Started (interval: 1m0s)
```

### 步骤 4: 配置 Sub2API 上游

#### 4.1 访问管理面板

打开浏览器访问：`http://23.80.88.63:12321/admin/upstreams`

#### 4.2 添加上游

点击"添加上游"按钮，填写以下信息：

| 字段 | 值 | 说明 |
|-----|-----|-----|
| **名称** | Sub2API Provider | 任意名称 |
| **Base URL** | `https://your-sub2api.com/openai` | ⚠️ 不要包含 `/v1` |
| **API Key** | `sk-xxx...` | 从 Sub2API 获取 |
| **优先级** | `0` | 数字越小优先级越高 |
| **状态** | **启用** | 必须选择"启用" |
| **权重** | `1` | 负载均衡权重 |
| **最大重试** | `3` | 失败重试次数 |
| **超时** | `120` | 请求超时（秒） |

点击"保存"。

#### 4.3 测试健康检查

点击"测试健康"按钮，应该看到：
- 蓝色提示框显示"正在检测 1/1 个上游..."
- 如果配置正确，上游状态应该显示为"正常"（绿色）
- 如果配置错误，状态会显示为"异常"（红色）

### 步骤 5: 创建 API Key

#### 5.1 访问仪表板

访问：`http://23.80.88.63:12321/dashboard`

#### 5.2 创建 API Key

1. 点击"API Keys"标签
2. 点击"创建 API Key"
3. 填写名称（例如："Codex Client"）
4. 点击"创建"
5. **复制并保存 API Key**（只显示一次）

### 步骤 6: 测试 Codex API

#### 6.1 测试 /v1/responses 端点

```bash
curl -X POST http://23.80.88.63:12321/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_GATEWAY_API_KEY" \
  -d '{
    "model": "gpt-5.1-codex",
    "input": [
      {
        "type": "message",
        "role": "user",
        "content": "Hello, test"
      }
    ],
    "stream": false
  }'
```

**预期响应**：
```json
{
  "id": "resp_xxx",
  "object": "response",
  "created": 1234567890,
  "model": "gpt-5.1-codex",
  "choices": [...],
  "usage": {
    "input_tokens": 10,
    "output_tokens": 20
  }
}
```

#### 6.2 测试流式请求

```bash
curl -X POST http://23.80.88.63:12321/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_GATEWAY_API_KEY" \
  -d '{
    "model": "gpt-5.1-codex",
    "input": [
      {
        "type": "message",
        "role": "user",
        "content": "Hello"
      }
    ],
    "stream": true
  }'
```

**预期响应**：SSE 流式数据
```
data: {"type":"response.started",...}

data: {"type":"response.content_part.delta",...}

data: {"type":"response.completed",...}
```

### 步骤 7: 配置 Codex 客户端

如果您使用 Codex CLI 或其他客户端：

```bash
# 设置环境变量
export OPENAI_API_BASE="http://23.80.88.63:12321"
export OPENAI_API_KEY="YOUR_GATEWAY_API_KEY"

# 或者在配置文件中设置
# ~/.codex/config.toml
[api]
base_url = "http://23.80.88.63:12321"
api_key = "YOUR_GATEWAY_API_KEY"
```

## 监控和日志

### 查看实时日志

```bash
# 所有日志
docker-compose logs -f backend

# 只看健康检查
docker-compose logs -f backend | grep "HealthCheck"

# 只看代理请求
docker-compose logs -f backend | grep "ProxyHandler"

# 只看上游选择
docker-compose logs -f backend | grep "Upstream"
```

### 查看健康状态

```bash
# 通过 API 查看
curl http://23.80.88.63:12321/api/admin/codex/upstreams/health \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### 查看使用统计

访问：`http://23.80.88.63:12321/dashboard`

- 查看 API 调用次数
- 查看 token 使用量
- 查看余额变化

## 故障排查

### 问题 1: 404 Not Found

**症状**：
```json
{"error": "404 page not found"}
```

**原因**：路径配置错误

**解决**：
1. 检查 Base URL 是否正确：`https://your-sub2api.com/openai`
2. 不要包含 `/v1` 在 Base URL 中
3. 网关会自动添加 `/v1/responses`

### 问题 2: 401 Unauthorized

**症状**：
```json
{"error": {"type": "authentication_error", "message": "Invalid API key"}}
```

**原因**：API Key 错误

**解决**：
1. 检查 Sub2API 的 API Key 是否正确
2. 在 Sub2API 管理面板确认 API Key 状态
3. 确认 API Key 没有过期

### 问题 3: 503 Service Unavailable

**症状**：
```json
{"error": "no available upstream"}
```

**原因**：没有启用的上游

**解决**：
1. 访问 `/admin/upstreams`
2. 确认至少有一个上游状态为"启用"
3. 点击"测试健康"检查连接

### 问题 4: 没有看到请求到达 Sub2API

**检查清单**：

1. **检查上游状态**：
   ```bash
   docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway \
     -c "SELECT id, name, status FROM codex_upstreams;"
   ```

2. **检查网关日志**：
   ```bash
   docker-compose logs backend | grep -E "ProxyHandler|Upstream|HealthCheck"
   ```

3. **手动触发健康检查**：
   ```bash
   curl -X POST http://localhost:8080/api/admin/codex/upstreams/health/check \
     -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
   ```

4. **查看详细日志**：
   ```bash
   docker-compose logs backend --tail=100
   ```

## 性能优化建议

### 1. 调整健康检查间隔

如果需要更频繁的健康检查，修改 `internal/upstream/health_checker.go`:

```go
checkInterval: 30 * time.Second,  // 改为 30 秒
```

### 2. 调整超时设置

在上游配置中调整超时时间（默认 120 秒）。

### 3. 添加更多上游

如果有多个 Sub2API 实例，可以添加多个上游实现负载均衡。

## 安全建议

### 1. 使用 HTTPS

在生产环境中，建议使用 Nginx 反向代理并启用 HTTPS：

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:12321;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 2. 限制访问

使用防火墙限制只允许特定 IP 访问：

```bash
# 只允许特定 IP
ufw allow from YOUR_IP to any port 12321

# 或使用 Nginx 限制
allow YOUR_IP;
deny all;
```

### 3. 定期更新

```bash
# 定期拉取更新
cd /root/codex-gateway
git pull origin main
./deploy-auto.sh
```

## 支持

如果遇到问题：

1. 查看日志：`docker-compose logs backend`
2. 检查配置：访问 `/admin/upstreams`
3. 测试连接：点击"测试健康"
4. 查看文档：`CODEX_SETUP_GUIDE.md`

## 总结

✅ **已完成**：
- Codex API 完整支持
- 多端点路由
- 健康检查和故障转移
- 用户会话亲和性
- 完整的计费系统
- 管理界面

🚀 **可以使用了**：
- 配置上游
- 创建 API Key
- 开始使用 Codex API

📊 **监控**：
- 查看日志
- 检查健康状态
- 监控使用量

祝使用愉快！
