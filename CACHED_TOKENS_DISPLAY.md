# 🎉 缓存Token显示功能已添加

## 更新内容

### 前端界面更新

在使用记录页面（`/usage`）中添加了**缓存Token**列：

| 时间 | 模型 | 输入Token | 输出Token | **缓存Token** | 总Token | 费用 | 延迟 | 状态 |
|------|------|-----------|-----------|--------------|---------|------|------|------|
| ... | gpt-5.2-codex | 4483 | 15 | **0** | 4498 | $0.008 | 4936ms | 200 |

### 视觉效果

- **有缓存时**: 显示为绿色高亮数字（例如：<span style="color: green; font-weight: bold;">1234</span>）
- **无缓存时**: 显示为灰色数字（例如：<span style="color: gray;">0</span>）

### 技术实现

1. **类型定义更新** (`frontend/src/types/api.ts`)
   ```typescript
   export interface UsageLog {
     // ...
     cached_tokens: number;  // 新增字段
     // ...
   }
   ```

2. **UI组件更新** (`frontend/src/app/(dashboard)/usage/page.tsx`)
   - 添加"缓存Token"表头
   - 添加缓存token数据列
   - 条件样式：缓存>0时显示绿色，否则显示灰色

### 后端支持

后端已经在存储和返回缓存token数据：

- **数据库字段**: `usage_logs.cached_tokens`
- **API响应**: 包含 `cached_tokens` 字段
- **计费逻辑**: 使用 `cache_read_price_per_1k` 计算缓存token费用

---

## 🚀 部署更新

### 1. 部署到服务器

```bash
ssh root@23.80.88.63
cd /root/codex-gateway
git pull origin main
./deploy-auto.sh
```

### 2. 验证显示

部署完成后：

1. 访问 https://api.codex-gateway.com/usage
2. 查看使用记录表格
3. 确认"缓存Token"列已显示

### 3. 测试缓存功能

发起一个带缓存的请求（需要重复相同的prompt）：

```bash
# 第一次请求（无缓存）
curl -X POST https://api.codex-gateway.com/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-5.2-codex",
    "messages": [{"role": "user", "content": "测试缓存功能"}],
    "stream": true
  }'

# 第二次请求（应该有缓存）
curl -X POST https://api.codex-gateway.com/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-5.2-codex",
    "messages": [{"role": "user", "content": "测试缓存功能"}],
    "stream": true
  }'
```

第二次请求应该显示缓存token数量（绿色高亮）。

---

## 📊 缓存Token说明

### 什么是缓存Token？

当您发送相似或重复的请求时，Codex API会缓存之前的输入内容，减少重复处理：

- **缓存命中**: 相同的输入内容会被缓存
- **费用优惠**: 缓存token的费用是正常输入token的10%
- **性能提升**: 缓存命中可以减少延迟

### 定价对比

以gpt-5.2-codex为例：

| Token类型 | 价格（per 1K） | 说明 |
|-----------|---------------|------|
| 输入Token | $0.00175 | 正常输入 |
| 输出Token | $0.014 | AI生成的内容 |
| **缓存Token** | **$0.000175** | **输入价格的10%** |

### 示例计算

假设一个请求：
- 输入: 1000 tokens
- 输出: 100 tokens
- 缓存: 500 tokens（来自之前的请求）

**费用计算**：
```
输入费用:  (1000 / 1000) × $0.00175 = $0.00175
缓存费用:  (500 / 1000) × $0.000175 = $0.0000875
输出费用:  (100 / 1000) × $0.014 = $0.0014
总费用: $0.00175 + $0.0000875 + $0.0014 = $0.0032375
```

如果没有缓存，费用会是：
```
输入费用:  (1500 / 1000) × $0.00175 = $0.002625
输出费用:  (100 / 1000) × $0.014 = $0.0014
总费用: $0.002625 + $0.0014 = $0.004025
```

**节省**: $0.004025 - $0.0032375 = $0.0007875（约19.6%）

---

## ✅ 更新清单

- [x] 后端已支持cached_tokens字段
- [x] 前端类型定义已更新
- [x] 使用记录表格已添加缓存列
- [x] 缓存token显示样式已优化
- [x] 代码已提交并推送

**现在可以部署了！**
