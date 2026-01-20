# Codex API 响应格式测试

## 问题
我们的网关显示：1252 输入，11273 输出
Sub2API 显示：11771 输入，306 输出

数据完全相反！

## 可能的原因

### 假设 1：Codex API 的字段含义
Codex API 可能使用了不同的命名约定：
- `input_tokens` = 用户输入的 tokens（prompt）
- `output_tokens` = AI 生成的 tokens（completion）

这是标准的理解。

### 假设 2：我们的解析有问题
让我们检查实际的 SSE 响应：

```
data: {"type":"response.completed","response":{"usage":{"input_tokens":X,"output_tokens":Y}}}
```

我们的代码：
```go
lastUsage.PromptTokens = codexEvent.Response.Usage.InputTokens  // X
lastUsage.CompletionTokens = codexEvent.Response.Usage.OutputTokens  // Y
```

然后存储：
```go
InputTokens:  inputTokens,   // X
OutputTokens: outputTokens,  // Y
```

### 假设 3：Sub2API 做了转换
可能 Sub2API 在某个地方交换了这两个值？

## 测试方法

在服务器上添加日志来查看实际的 API 响应：

```go
// 在 proxy.go 的 246 行后添加
if err := json.Unmarshal([]byte(data), &codexEvent); err == nil && codexEvent.Type == "response.completed" {
    log.Printf("[DEBUG] Codex response.completed: input_tokens=%d, output_tokens=%d, cached_tokens=%d",
        codexEvent.Response.Usage.InputTokens,
        codexEvent.Response.Usage.OutputTokens,
        codexEvent.Response.Usage.InputTokenDetails.CachedTokens)

    lastUsage.PromptTokens = codexEvent.Response.Usage.InputTokens
    lastUsage.CompletionTokens = codexEvent.Response.Usage.OutputTokens
    lastUsage.CachedTokens = codexEvent.Response.Usage.InputTokenDetails.CachedTokens
    lastUsage.TotalTokens = codexEvent.Response.Usage.InputTokens + codexEvent.Response.Usage.OutputTokens
    continue
}
```

## 下一步

1. 添加调试日志
2. 发起一个测试请求
3. 查看日志中的实际值
4. 对比 Sub2API 的显示

这样我们就能确定问题所在。
