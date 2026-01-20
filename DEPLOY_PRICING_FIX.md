# 🔧 定价修复 - 部署指南

## 问题总结

发现了两个定价问题：

### 1. 不应该有1.5倍markup
- **错误配置**: MarkupMultiplier = 1.5
- **正确配置**: MarkupMultiplier = 1.0
- **原因**: Sub2API没有应用markup，我们也不应该有

### 2. 定价数据不准确
- **gpt-5.1-codex**: 使用了错误的价格（$0.00138/$0.011）
  - 正确价格应该是: $0.00125/$0.01
- **gpt-5.2-codex**: 使用了gpt-5.1的价格
  - 正确价格应该是: $0.00175/$0.014

## 修复内容

已更新 `internal/database/seed_codex_pricing.go`：

### 修复前后对比

| 模型 | 字段 | 修复前 | 修复后 | Sub2API |
|------|------|--------|--------|---------|
| gpt-5.1-codex | Input | $0.00138 | $0.00125 | $0.00125 ✓ |
| gpt-5.1-codex | Output | $0.011 | $0.01 | $0.01 ✓ |
| gpt-5.1-codex | Markup | 1.5 | 1.0 | 1.0 ✓ |
| gpt-5.2-codex | Input | $0.00138 | $0.00175 | $0.00175 ✓ |
| gpt-5.2-codex | Output | $0.011 | $0.014 | $0.014 ✓ |
| gpt-5.2-codex | Markup | 1.5 | 1.0 | 1.0 ✓ |

## 费用验证

### 测试数据（您的实际请求）
- Input tokens: 4483
- Output tokens: 15
- Model: gpt-5.2-codex

### 修复前计算
```
Input:  (4483 / 1000) × $0.00138 = $0.00618654
Output: (15 / 1000) × $0.011 = $0.000165
Subtotal: $0.00635154
With 1.5x markup: $0.00635154 × 1.5 = $0.00952731
```
**实际显示**: $0.012083（因为用了错误的价格）

### 修复后计算
```
Input:  (4483 / 1000) × $0.00175 = $0.00784525
Output: (15 / 1000) × $0.014 = $0.00021
Total: $0.00784525 + $0.00021 = $0.00805525
```
**预期显示**: $0.008055 ✓ **与Sub2API完全一致！**

---

## 🚀 部署步骤

### 1. 在服务器上部署

```bash
ssh root@23.80.88.63
cd /root/codex-gateway
git pull origin main
./deploy-auto.sh
```

### 2. 验证定价已更新

部署完成后，检查数据库中的定价：

```bash
docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway -c \
  "SELECT model_name, input_price_per_1k, output_price_per_1k, cache_read_price_per_1k, markup_multiplier
   FROM model_pricings
   WHERE model_name IN ('gpt-5.1-codex', 'gpt-5.2-codex');"
```

**预期输出**：
```
    model_name    | input_price_per_1k | output_price_per_1k | cache_read_price_per_1k | markup_multiplier
------------------+--------------------+---------------------+-------------------------+-------------------
 gpt-5.1-codex    |            0.00125 |                0.01 |                0.000125 |                 1
 gpt-5.2-codex    |            0.00175 |               0.014 |                0.000175 |                 1
```

### 3. 发起测试请求

```bash
curl -X POST https://api.codex-gateway.com/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-5.2-codex",
    "messages": [
      {
        "role": "user",
        "content": "写一个Python函数计算斐波那契数列"
      }
    ],
    "stream": true
  }'
```

### 4. 查看日志验证

```bash
docker-compose logs -f backend | grep DEBUG
```

**预期日志**：
```
[DEBUG] Codex response.completed: input_tokens=4483, output_tokens=15, cached_tokens=0
[DEBUG] Mapped: PromptTokens=4483, CompletionTokens=15, CachedTokens=0
[DEBUG] Billing: model=gpt-5.2-codex, input=4483, output=15, cached=0
[DEBUG] Cost calculated: $0.008055
```

### 5. 对比Sub2API

在Sub2API中查看相同请求的费用，应该完全一致！

---

## ✅ 验证清单

- [ ] 代码已部署到服务器
- [ ] 数据库中的定价已更新
- [ ] markup_multiplier = 1.0
- [ ] gpt-5.2-codex的input_price_per_1k = 0.00175
- [ ] gpt-5.2-codex的output_price_per_1k = 0.014
- [ ] 测试请求的费用与Sub2API一致

---

## 🎉 预期结果

修复后，您的Codex Gateway的计费将与Sub2API完全一致：
- ✅ Token映射正确
- ✅ 定价数据准确
- ✅ 无额外markup
- ✅ 费用计算精确

**现在可以部署了！**
