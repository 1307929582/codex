# Codex Billing Implementation Test

## Test Case 1: Streaming Request with Cache

### Input (SSE Stream)
```
data: {"type":"response.started","id":"resp_123"}

data: {"type":"response.content_part.delta","delta":{"type":"text","text":"Hello"}}

data: {"type":"response.completed","response":{"id":"resp_123","usage":{"input_tokens":15,"output_tokens":4463,"input_tokens_details":{"cached_tokens":2650}}}}

data: [DONE]
```

### Expected Parsing
- `lastUsage.PromptTokens` = 15
- `lastUsage.CompletionTokens` = 4463
- `lastUsage.CachedTokens` = 2650
- `lastUsage.TotalTokens` = 15 + 4463 = 4478

### Expected Billing (gpt-5.2-codex)
```
Pricing:
- InputPricePer1k: $0.00138
- CacheReadPricePer1k: $0.000138 (10% of input)
- OutputPricePer1k: $0.011
- MarkupMultiplier: 1.5

Calculation:
- inputCost = (15 / 1000) × $0.00138 = $0.0000207
- cacheReadCost = (2650 / 1000) × $0.000138 = $0.0003657
- outputCost = (4463 / 1000) × $0.011 = $0.0490930
- subtotal = $0.0494794
- total = $0.0494794 × 1.5 = $0.0742191
```

### Sub2API Comparison
```
Sub2API shows: $0.008020

Difference: $0.0742191 vs $0.008020 = 9.25x

Possible reasons:
1. Sub2API has lower base pricing
2. Sub2API has lower markup (maybe 1.0x instead of 1.5x)
3. Sub2API has different cache read pricing
```

## Test Case 2: Non-Streaming Request

### Input (JSON Response)
```json
{
  "id": "resp_456",
  "usage": {
    "input_tokens": 20,
    "output_tokens": 100,
    "input_tokens_details": {
      "cached_tokens": 50
    }
  }
}
```

### Expected Parsing
- `inputTokens` = 20
- `outputTokens` = 100
- `cachedTokens` = 50

### Expected Billing
```
- inputCost = (20 / 1000) × $0.00138 = $0.0000276
- cacheReadCost = (50 / 1000) × $0.000138 = $0.0000069
- outputCost = (100 / 1000) × $0.011 = $0.0011
- subtotal = $0.0011345
- total = $0.0011345 × 1.5 = $0.00170175
```

## Code Review Checklist

### ✅ Streaming Request Handling
- [x] Parse `response.completed` event
- [x] Extract `event.response.usage.input_tokens`
- [x] Extract `event.response.usage.output_tokens`
- [x] Extract `event.response.usage.input_tokens_details.cached_tokens`
- [x] Calculate `TotalTokens` correctly
- [x] Fallback to estimation if usage missing

### ✅ Non-Streaming Request Handling
- [x] Parse `usage` from response body
- [x] Extract `usage.input_tokens`
- [x] Extract `usage.output_tokens`
- [x] Extract `usage.input_tokens_details.cached_tokens`

### ✅ Billing Calculation
- [x] Separate costs for input, cache read, output
- [x] Correct unit conversion (per 1K tokens)
- [x] Apply markup multiplier
- [x] Handle zero cached tokens

### ✅ Database Recording
- [x] Store all three token types
- [x] Store calculated cost
- [x] Update user balance
- [x] Create usage log

## Potential Issues

### Issue 1: Markup Difference
**Problem**: Our markup (1.5x) may be higher than Sub2API's
**Solution**: Check if Sub2API uses different markup for different models

### Issue 2: Base Pricing
**Problem**: Our base pricing may be higher
**Solution**: Verify against LiteLLM pricing data

### Issue 3: Cache Read Pricing
**Problem**: 10% may not be correct for all models
**Solution**: Check LiteLLM data for actual cache_read_input_token_cost

## Next Steps

1. ✅ Deploy current implementation
2. ⏳ Test with real requests
3. ⏳ Compare actual billing with Sub2API
4. ⏳ Adjust pricing if needed
