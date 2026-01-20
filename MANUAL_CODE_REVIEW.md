# Codex Billing Code Review - Manual Analysis

## 代码审查结果

### ✅ 正确的实现

#### 1. SSE 事件解析（proxy.go:232-252）
```go
// ✓ 正确：优先解析 Codex response.completed 事件
var codexEvent struct {
    Type     string `json:"type"`
    Response struct {
        Usage struct {
            InputTokens       int `json:"input_tokens"`
            OutputTokens      int `json:"output_tokens"`
            InputTokenDetails struct {
                CachedTokens int `json:"cached_tokens"`
            } `json:"input_tokens_details"`
        } `json:"usage"`
    } `json:"response"`
}

if err := json.Unmarshal([]byte(data), &codexEvent); err == nil && codexEvent.Type == "response.completed" {
    lastUsage.PromptTokens = codexEvent.Response.Usage.InputTokens
    lastUsage.CompletionTokens = codexEvent.Response.Usage.OutputTokens
    lastUsage.CachedTokens = codexEvent.Response.Usage.InputTokenDetails.CachedTokens
    lastUsage.TotalTokens = codexEvent.Response.Usage.InputTokens + codexEvent.Response.Usage.OutputTokens
    continue  // ✓ 正确：跳过后续解析
}
```

**评价**：✅ 完全正确，与 Sub2API 实现一致

#### 2. ChatGPT 格式 Fallback（proxy.go:255-269）
```go
// ✓ 正确：如果不是 Codex 事件，尝试 ChatGPT 格式
var chunk OpenAIResponse
if err := json.Unmarshal([]byte(data), &chunk); err == nil {
    if chunk.Usage.TotalTokens > 0 {
        // ChatGPT 格式
        lastUsage.PromptTokens = chunk.Usage.PromptTokens
        lastUsage.CompletionTokens = chunk.Usage.CompletionTokens
        lastUsage.TotalTokens = chunk.Usage.TotalTokens
    } else if chunk.Usage.InputTokens > 0 || chunk.Usage.OutputTokens > 0 {
        // 直接 usage 格式（非事件）
        lastUsage.PromptTokens = chunk.Usage.InputTokens
        lastUsage.CompletionTokens = chunk.Usage.OutputTokens
        lastUsage.CachedTokens = chunk.Usage.InputTokenDetails.CachedTokens
        lastUsage.TotalTokens = chunk.Usage.InputTokens + chunk.Usage.OutputTokens
    }
}
```

**评价**：✅ 正确，支持多种格式

#### 3. 计费计算（proxy.go:385-398）
```go
func calculateCostWithCache(model string, inputTokens, outputTokens, cachedTokens int) (float64, error) {
    var pricing models.ModelPricing
    if err := database.DB.Where("model_name = ?", model).First(&pricing).Error; err != nil {
        return 0, fmt.Errorf("pricing not found for model: %s", model)
    }

    // ✓ 正确：分别计算三种 token 类型
    inputCost := (float64(inputTokens) / 1000.0) * pricing.InputPricePer1k
    cacheReadCost := (float64(cachedTokens) / 1000.0) * pricing.CacheReadPricePer1k
    outputCost := (float64(outputTokens) / 1000.0) * pricing.OutputPricePer1k

    // ✓ 正确：应用 markup
    return (inputCost + cacheReadCost + outputCost) * pricing.MarkupMultiplier, nil
}
```

**评价**：✅ 逻辑完全正确

#### 4. Fallback 估算（proxy.go:287-298）
```go
} else if streamedChunks > 0 {
    // ✓ 正确：当无法获取 usage 时使用估算
    estimatedTokens := streamedChunks * 25
    estimatedInput := estimatedTokens / 10
    estimatedOutput := estimatedTokens - estimatedInput

    cost, err := calculateCostWithCache(model, estimatedInput, estimatedOutput, 0)
    if err == nil {
        _ = recordUsageAndBill(user.ID, apiKey.ID, model, estimatedInput, estimatedOutput, 0, cost, latencyMs)
    }
}
```

**评价**：✅ 合理的 fallback 机制

### ⚠️ 潜在问题

#### 问题 1：定价过高
**位置**：`internal/database/seed_codex_pricing.go`

**当前定价**：
```go
InputPricePer1k:     0.00138   // $1.38 per 1M tokens
CacheReadPricePer1k: 0.000138  // $0.138 per 1M tokens
OutputPricePer1k:    0.011     // $11 per 1M tokens
MarkupMultiplier:    1.5
```

**Sub2API 实际定价（反推）**：
```go
InputPricePer1k:     0.000224  // $0.224 per 1M tokens
CacheReadPricePer1k: 0.0000224 // $0.0224 per 1M tokens
OutputPricePer1k:    0.001784  // $1.784 per 1M tokens
MarkupMultiplier:    1.0 (推测)
```

**差异**：我们的定价是 Sub2API 的 **6.2x**，加上 markup 差异，总计 **9.3x**

**建议**：调整定价以匹配 Sub2API

#### 问题 2：ChatGPT 格式的 CachedTokens
**位置**：`proxy.go:266`

**代码**：
```go
lastUsage.CachedTokens = chunk.Usage.InputTokenDetails.CachedTokens
```

**分析**：
- ChatGPT API 可能没有 `input_tokens_details` 字段
- 如果字段不存在，`CachedTokens` 会是 0（Go 的零值）
- 这是安全的，不会导致错误

**评价**：✅ 安全，无需修改

### 📊 测试结果

#### 实际请求测试
```
输入: 15 tokens
输出: 4463 tokens
缓存: 2650 tokens

我们的计费:
- inputCost = (15 / 1000) × $0.00138 = $0.0000207
- cacheReadCost = (2650 / 1000) × $0.000138 = $0.0003657
- outputCost = (4463 / 1000) × $0.011 = $0.0490930
- subtotal = $0.0494794
- total = $0.0494794 × 1.5 = $0.0742191

Sub2API 计费: $0.008020

差异: 9.25x
```

### 🎯 结论

**代码质量**：✅ 优秀
- 逻辑正确
- 结构清晰
- 错误处理完善
- 与 Sub2API 实现一致

**唯一问题**：定价配置过高

**建议修复**：调整 `seed_codex_pricing.go` 中的定价值

### 📝 推荐的定价调整

```diff
--- a/internal/database/seed_codex_pricing.go
+++ b/internal/database/seed_codex_pricing.go
@@ -13,9 +13,9 @@ func SeedCodexPricing() error {
 	codexModels := []models.ModelPricing{
 		{
 			ModelName:           "gpt-5.1-codex",
-			InputPricePer1k:     0.00138,
-			CacheReadPricePer1k: 0.000138,
-			OutputPricePer1k:    0.011,
+			InputPricePer1k:     0.000224,  // 降低到 Sub2API 水平
+			CacheReadPricePer1k: 0.0000224, // 10% of input
+			OutputPricePer1k:    0.001784,  // 降低到 Sub2API 水平
 			MarkupMultiplier:    1.5,
 		},
```

**效果**：费用将降至约 $0.012（比 Sub2API 高 50%，但可接受）
