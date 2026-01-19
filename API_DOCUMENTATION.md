# Codex Gateway API 文档

## 概述

Codex Gateway提供两套API：
1. **控制平面API** (`/api/*`) - 用户管理、API密钥管理、使用量查询
2. **数据平面API** (`/v1/*`) - OpenAI代理服务

---

## 认证方式

### 控制平面 - JWT认证
```
Authorization: Bearer <JWT_TOKEN>
```

### 数据平面 - API密钥认证
```
Authorization: Bearer <API_KEY>
```

---

## 控制平面API

### 1. 用户认证

#### 1.1 用户注册
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"  // 最少8位
}
```

**响应**:
```json
{
  "message": "user registered successfully"
}
```

#### 1.2 用户登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "user@example.com"
  }
}
```

#### 1.3 获取当前用户信息
```http
GET /api/auth/me
Authorization: Bearer <JWT_TOKEN>
```

**响应**:
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "balance": 100.50,
  "status": "active",
  "created_at": "2026-01-19T10:00:00Z"
}
```

---

### 2. API密钥管理

#### 2.1 获取密钥列表
```http
GET /api/keys
Authorization: Bearer <JWT_TOKEN>
```

**响应**:
```json
[
  {
    "id": 1,
    "user_id": "uuid",
    "key_prefix": "sk-abc",
    "name": "Production Key",
    "quota_limit": null,
    "total_usage": 15000,
    "status": "active",
    "created_at": "2026-01-19T10:00:00Z",
    "last_used_at": "2026-01-19T12:00:00Z"
  }
]
```

#### 2.2 创建新密钥
```http
POST /api/keys
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "name": "My API Key",
  "quota_limit": 100.0  // 可选，限制此密钥的最大消费
}
```

**响应**:
```json
{
  "id": 1,
  "key": "sk-abc123def456...",  // 完整密钥，仅显示一次！
  "name": "My API Key"
}
```

⚠️ **重要**: 完整密钥仅在创建时返回一次，请妥善保存！

#### 2.3 删除密钥
```http
DELETE /api/keys/:id
Authorization: Bearer <JWT_TOKEN>
```

**响应**:
```json
{
  "message": "key deleted"
}
```

#### 2.4 更新密钥状态
```http
PUT /api/keys/:id/status
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "status": "disabled"  // active 或 disabled
}
```

**响应**:
```json
{
  "message": "status updated"
}
```

---

### 3. 使用量查询

#### 3.1 获取使用日志
```http
GET /api/usage/logs?page=1&page_size=20&model=gpt-4
Authorization: Bearer <JWT_TOKEN>
```

**查询参数**:
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20，最大100）
- `model`: 筛选模型（可选）

**响应**:
```json
{
  "data": [
    {
      "request_id": "uuid",
      "user_id": "uuid",
      "api_key_id": 1,
      "model": "gpt-4",
      "input_tokens": 100,
      "output_tokens": 200,
      "total_tokens": 300,
      "cost": 0.015,
      "latency_ms": 1500,
      "status_code": 200,
      "created_at": "2026-01-19T12:00:00Z"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 150,
  "total_page": 8
}
```

#### 3.2 获取使用统计
```http
GET /api/usage/stats
Authorization: Bearer <JWT_TOKEN>
```

**响应**:
```json
{
  "today_cost": 5.25,
  "month_cost": 125.50,
  "total_cost": 1250.00
}
```

---

### 4. 账户管理

#### 4.1 获取余额
```http
GET /api/account/balance
Authorization: Bearer <JWT_TOKEN>
```

**响应**:
```json
{
  "balance": 100.50,
  "currency": "USD"
}
```

#### 4.2 获取交易记录
```http
GET /api/account/transactions
Authorization: Bearer <JWT_TOKEN>
```

**响应**:
```json
[
  {
    "id": "uuid",
    "user_id": "uuid",
    "amount": 100.00,
    "type": "deposit",
    "description": "Initial deposit",
    "created_at": "2026-01-19T10:00:00Z"
  }
]
```

---

## 数据平面API

### OpenAI代理

#### Chat Completions
```http
POST /v1/chat/completions
Authorization: Bearer <API_KEY>
Content-Type: application/json

{
  "model": "gpt-3.5-turbo",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "temperature": 0.7,
  "max_tokens": 100
}
```

**响应**: 与OpenAI API完全兼容

---

## 错误响应

所有错误响应格式：
```json
{
  "error": "error message"
}
```

### HTTP状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或认证失败 |
| 402 | 余额不足 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 502 | 上游服务错误 |

---

## 使用示例

### 完整流程示例

```bash
# 1. 注册用户
curl -X POST http://localhost:12322/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 2. 登录获取JWT
TOKEN=$(curl -X POST http://localhost:12322/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

# 3. 创建API密钥
API_KEY=$(curl -X POST http://localhost:12322/api/keys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Key"}' \
  | jq -r '.key')

# 4. 使用API密钥调用OpenAI代理
curl -X POST http://localhost:12322/v1/chat/completions \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# 5. 查看使用统计
curl -X GET http://localhost:12322/api/usage/stats \
  -H "Authorization: Bearer $TOKEN"
```

---

## 速率限制

当前版本暂无速率限制，将在Phase 3实现。

---

## 安全建议

1. **JWT Token**: 有效期24小时，请妥善保存
2. **API密钥**: 仅在创建时显示一次，请立即保存
3. **HTTPS**: 生产环境务必使用HTTPS
4. **JWT_SECRET**: 生产环境必须修改默认值

---

## 更新日志

### v1.0.0 (2026-01-19)
- ✅ 用户认证系统（JWT）
- ✅ API密钥管理
- ✅ 使用量查询
- ✅ 账户余额管理
- ✅ OpenAI代理服务
- ✅ 原子化计费系统
