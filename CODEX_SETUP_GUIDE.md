# Codex Gateway 配置指南

## 架构说明

```
Codex Client → 本网关 → Sub2API → ChatGPT/OpenAI
```

- **本网关**：接收 Codex 请求，转发到 Sub2API
- **Sub2API**：管理 ChatGPT OAuth 账户，处理实际 API 调用

## 配置步骤

### 1. 添加 Sub2API 上游

访问 `http://YOUR_SERVER:12321/admin/upstreams`，点击"添加上游"：

**配置示例**：
- **名称**: Sub2API Provider
- **Base URL**: `https://your-sub2api-domain.com/openai` ⚠️ **注意：不要包含 /v1**
- **API Key**: 您的 Sub2API API Key（从 Sub2API 管理面板获取）
- **优先级**: 0
- **状态**: **启用**
- **权重**: 1
- **最大重试**: 3
- **超时**: 120

### 2. 路径映射

本网关会自动处理路径映射：

| 客户端请求路径 | 转发到 Sub2API |
|--------------|---------------|
| `/v1/responses` | `{base_url}/v1/responses` |
| `/v1/completions` | `{base_url}/v1/completions` |
| `/v1/chat/completions` | `{base_url}/v1/chat/completions` |

**示例**：
- 客户端请求：`http://YOUR_SERVER:12321/v1/responses`
- 转发到：`https://your-sub2api-domain.com/openai/v1/responses`

### 3. Codex 客户端配置

配置您的 Codex 客户端：

```bash
# 设置 Base URL
export OPENAI_API_BASE="http://YOUR_SERVER:12321"

# 设置 API Key（从本网关管理面板创建）
export OPENAI_API_KEY="your-gateway-api-key"
```

### 4. 测试连接

```bash
curl -X POST http://YOUR_SERVER:12321/v1/responses \
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
    "stream": false
  }'
```

## 重要说明

### ✅ 本网关会做的事情

1. 接收 Codex API 请求（`/v1/responses` 等）
2. 验证用户 API Key 和余额
3. 选择上游（用户会话亲和性）
4. 转发请求到 Sub2API
5. 返回响应并计费

### ❌ 本网关不会做的事情

1. **不修改请求体**：Codex 请求原样转发（Sub2API 会处理转换）
2. **不添加特殊 headers**：OAuth 相关 headers 由 Sub2API 处理
3. **不处理 instructions**：Sub2API 会自动添加 Codex instructions

### 与 ChatGPT API 的区别

| 特性 | ChatGPT API | Codex API |
|-----|------------|-----------|
| 端点 | `/v1/chat/completions` | `/v1/responses` |
| 请求字段 | `messages` | `input` |
| 转换 | 应用 ChatGPT 转换 | 原样转发 |
| stream_options | 添加 | 不添加 |

## 故障排查

### 问题：404 Not Found

**原因**：路径配置错误

**解决**：
- 确保 Base URL 是 `https://your-sub2api.com/openai`（不要包含 `/v1`）
- 本网关会自动添加 `/v1/responses` 等路径

### 问题：401 Unauthorized

**原因**：API Key 错误

**解决**：
- 检查 Sub2API 的 API Key 是否正确
- 在 Sub2API 管理面板确认 API Key 状态

### 问题：没有看到请求

**原因**：上游状态为 disabled

**解决**：
- 在 `/admin/upstreams` 确认上游状态为"启用"
- 点击"测试健康"检查连接

## 查看日志

```bash
# 查看网关日志
docker-compose logs -f backend

# 查看健康检查日志
docker-compose logs -f backend | grep "HealthCheck"

# 查看代理请求日志
docker-compose logs -f backend | grep "ProxyHandler"
```

## 计费说明

- 本网关根据 Sub2API 返回的 token 使用量计费
- 流式请求：从 SSE 响应中解析 usage 信息
- 非流式请求：从响应 JSON 中获取 usage 信息
- 如果 usage 信息缺失，使用估算值（流式：每个 chunk 约 25 tokens）

## 会话管理

本网关使用一致性哈希确保：
- **同一用户**的请求始终路由到**同一上游**
- 保持 Sub2API 的会话缓存有效
- 避免频繁切换导致上下文丢失
