# 🚀 部署和测试指南

## 第一步：部署到服务器

```bash
ssh root@23.80.88.63
cd /root/codex-gateway
git pull origin main
./deploy-auto.sh
```

等待部署完成（大约1-2分钟）。

---

## 第二步：发起测试请求

使用您的API密钥发起一个Codex API请求：

```bash
curl -X POST https://api.codex-gateway.com/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "gpt-5.1-codex",
    "messages": [
      {
        "role": "user",
        "content": "写一个Python函数计算斐波那契数列"
      }
    ],
    "stream": true
  }'
```

**注意**：
- 替换 `YOUR_API_KEY` 为您的实际API密钥
- 使用 `stream: true` 来触发流式响应的代码路径
- 这会触发我们添加的调试日志

---

## 第三步：查看调试日志

在服务器上运行：

```bash
docker-compose logs -f backend | grep DEBUG
```

或者查看最近100行日志：

```bash
docker-compose logs --tail=100 backend | grep DEBUG
```

---

## 📊 预期日志输出

您应该看到类似这样的输出：

```
[DEBUG] Codex response.completed: input_tokens=1252, output_tokens=11273, cached_tokens=0
[DEBUG] Mapped: PromptTokens=1252, CompletionTokens=11273, CachedTokens=0
[DEBUG] Billing: model=gpt-5.1-codex, input=1252, output=11273, cached=0
[DEBUG] Cost calculated: $0.xxxxxx
```

---

## 🔍 关键信息解读

### 1. **Codex response.completed**
这是Codex API返回的**原始值**：
- `input_tokens`: Codex API返回的输入token数
- `output_tokens`: Codex API返回的输出token数
- `cached_tokens`: Codex API返回的缓存token数

### 2. **Mapped**
这是我们**映射后**的值：
- `PromptTokens`: 映射到我们系统的输入token
- `CompletionTokens`: 映射到我们系统的输出token
- `CachedTokens`: 映射到我们系统的缓存token

### 3. **Billing**
这是**计费时**使用的值，应该与Mapped相同。

### 4. **Cost calculated**
这是**计算出的费用**。

---

## ❓ 诊断问题

### 场景A：数值与Sub2API相反

如果您看到：
```
[DEBUG] Codex response.completed: input_tokens=11771, output_tokens=306
```

但Sub2API显示：
- 输入: 306
- 输出: 11771

**结论**：Codex API的字段命名与我们的理解相反！

**解决方案**：需要交换映射逻辑。

---

### 场景B：数值一致但费用不对

如果token数量正确，但费用计算错误：

1. 检查定价配置：
```bash
docker exec -it codex-gateway-db-1 psql -U codex_user -d codex_gateway -c \
  "SELECT model_name, input_price_per_1k, output_price_per_1k, cache_read_price_per_1k, markup_multiplier FROM model_pricings WHERE model_name = 'gpt-5.1-codex';"
```

2. 手动验证计算：
```
inputCost = (input_tokens / 1000) × input_price_per_1k
cacheReadCost = (cached_tokens / 1000) × cache_read_price_per_1k
outputCost = (output_tokens / 1000) × output_price_per_1k
totalCost = (inputCost + cacheReadCost + outputCost) × markup_multiplier
```

---

### 场景C：没有看到DEBUG日志

如果没有看到任何DEBUG日志：

1. 确认部署成功：
```bash
docker-compose ps
```

2. 检查容器是否重启：
```bash
docker-compose logs backend | tail -50
```

3. 确认代码版本：
```bash
git log --oneline -1
```

应该看到：`8bb67dc Fix: Add missing log import`

---

## 📝 下一步

1. **执行部署**
2. **发起测试请求**
3. **复制完整的DEBUG日志输出**
4. **发送给我分析**

我会根据日志输出确定问题根源并提供修复方案！
