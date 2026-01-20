# 🔍 定价问题分析

## 问题发现

### Sub2API的gpt-5.2实际定价（per token）
```json
"gpt-5.2": {
    "input_cost_per_token": 1.75e-06,      // $0.00000175
    "output_cost_per_token": 1.4e-05,      // $0.000014
    "cache_read_input_token_cost": 1.75e-07 // $0.000000175
}
```

**转换为per 1K tokens**：
- Input: $0.00175 per 1K
- Output: $0.014 per 1K
- Cache Read: $0.000175 per 1K

### 我们系统当前配置（错误）
```go
{
    ModelName:           "gpt-5.2-codex",
    InputPricePer1k:     0.00138,  // ❌ 这是gpt-5.1的价格！
    OutputPricePer1k:    0.011,    // ❌ 这是gpt-5.1的价格！
    CacheReadPricePer1k: 0.000138, // ❌ 这是gpt-5.1的价格！
    MarkupMultiplier:    1.5,      // ❌ 不应该有markup
}
```

## 费用计算验证

### 测试数据
- Input tokens: 4483
- Output tokens: 15
- Cached tokens: 0

### 使用正确的gpt-5.2定价（无markup）
```
Input cost:  (4483 / 1000) × $0.00175 = $0.00784525
Output cost: (15 / 1000) × $0.014 = $0.00021
Total: $0.00784525 + $0.00021 = $0.00805525
```

**Sub2API显示**: $0.008055 ✓ **完全匹配！**

### 使用我们当前的错误配置
```
Input cost:  (4483 / 1000) × $0.00175 = $0.00784525
Output cost: (15 / 1000) × $0.014 = $0.00021
Total: ($0.00784525 + $0.00021) × 1.5 = $0.01208288
```

**我们系统显示**: $0.012083 ✓ **匹配我们的错误计算！**

## 问题根源

### 1. 定价数据错误
gpt-5.2-codex应该使用gpt-5.2的定价，而不是gpt-5.1的定价。

### 2. 不应该有markup
Sub2API没有应用markup，所以我们也不应该有。

## 修复方案

### 更新seed_codex_pricing.go

需要修改：
1. gpt-5.2-codex的定价改为gpt-5.2的正确价格
2. 所有模型的markup_multiplier改为1.0

### 正确的配置应该是：

```go
{
    ModelName:           "gpt-5.1-codex",
    InputPricePer1k:     0.00125,  // Sub2API: 1.25e-06 × 1000
    OutputPricePer1k:    0.01,     // Sub2API: 1e-05 × 1000
    CacheReadPricePer1k: 0.000125, // Sub2API: 1.25e-07 × 1000
    MarkupMultiplier:    1.0,      // 无加价
},
{
    ModelName:           "gpt-5.2-codex",
    InputPricePer1k:     0.00175,  // Sub2API: 1.75e-06 × 1000
    OutputPricePer1k:    0.014,    // Sub2API: 1.4e-05 × 1000
    CacheReadPricePer1k: 0.000175, // Sub2API: 1.75e-07 × 1000
    MarkupMultiplier:    1.0,      // 无加价
},
```

## 下一步

1. 修复seed_codex_pricing.go
2. 重新部署
3. 测试验证费用是否与Sub2API一致
