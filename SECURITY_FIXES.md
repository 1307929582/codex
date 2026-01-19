# 安全漏洞修复报告

## 修复日期
2026-01-19

## 审计来源
Gemini Code Audit

---

## 已修复的关键漏洞

### 1. JWT中间件DoS漏洞 ⚠️ CRITICAL

**文件**: `internal/middleware/jwt.go`

**问题描述**:
- 使用 `uuid.MustParse()` 处理JWT中的user_id
- 如果恶意用户在JWT中放入非UUID格式的字符串，会导致panic并使整个服务器崩溃

**修复方案**:
```go
// 修复前
userID := claims["user_id"].(string)
if err := database.DB.First(&user, "id = ?", uuid.MustParse(userID)).Error

// 修复后
userIDStr, ok := claims["user_id"].(string)
if !ok {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in token"})
    c.Abort()
    return
}

userID, err := uuid.Parse(userIDStr)
if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID format"})
    c.Abort()
    return
}
```

**影响**: 防止DoS攻击，提高系统稳定性

---

### 2. 财务漏洞 ⚠️ CRITICAL

**文件**: `internal/handlers/proxy.go`

**问题描述**:
- 余额检查发生在OpenAI API调用**之后**
- 零余额用户可以无限制调用OpenAI API，造成财务损失

**修复方案**:
```go
// 在调用OpenAI API之前添加余额预检查
if user.Balance <= 0 {
    c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
    return
}

startTime := time.Now()
upstreamResp, err := forwardToOpenAI(c.Request.Context(), req)
```

**影响**: 防止财务损失，确保计费系统正确性

---

### 3. 硬编码JWT密钥 ⚠️ CRITICAL

**文件**: `internal/config/config.go`

**问题描述**:
- JWT_SECRET有默认值 `"change-me-in-production"`
- 如果用户忘记设置环境变量，系统会使用不安全的默认值

**修复方案**:
```go
JWTSecret: getEnv("JWT_SECRET", ""),

// 添加验证
if AppConfig.JWTSecret == "" {
    log.Fatal("JWT_SECRET is required")
}

if len(AppConfig.JWTSecret) < 32 {
    log.Fatal("JWT_SECRET must be at least 32 characters")
}
```

**影响**: 强制使用强密钥，防止JWT伪造攻击

---

### 4. 缺少CORS配置 ⚠️ HIGH

**文件**: `cmd/gateway/main.go`

**问题描述**:
- 没有配置CORS中间件
- 前端无法正常调用API

**修复方案**:
```go
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))
```

**影响**: 允许前端正常访问API

---

### 5. 缺少优雅关闭 ⚠️ MEDIUM

**文件**: `cmd/gateway/main.go`

**问题描述**:
- 服务器不处理SIGINT/SIGTERM信号
- 强制关闭可能导致数据丢失或连接中断

**修复方案**:
```go
srv := &http.Server{
    Addr:    ":" + config.AppConfig.ServerPort,
    Handler: router,
}

go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal("Failed to start server:", err)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

log.Println("Shutting down server...")
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    log.Fatal("Server forced to shutdown:", err)
}
```

**影响**: 确保服务器可以优雅关闭，保护数据完整性

---

### 6. 前端XSS风险 ⚠️ HIGH

**文件**: `frontend/src/lib/stores/auth.ts`

**问题描述**:
- 直接使用 `localStorage.getItem()` 和 `localStorage.setItem()`
- 在SSR环境中会导致hydration错误
- JWT存储在localStorage容易受到XSS攻击

**修复方案**:
```typescript
// 使用Zustand persist中间件
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      setAuth: (token, user) => {
        set({ token, user });
      },
      logout: () => {
        set({ token: null, user: null });
      },
      isAuthenticated: () => !!get().token,
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ token: state.token, user: state.user }),
    }
  )
);
```

**影响**:
- 修复SSR hydration错误
- 添加安全警告注释
- 建议生产环境使用HttpOnly cookies

---

## 未修复的次要问题

### 1. Float64精度问题 ⚠️ MEDIUM

**问题**: 使用float64处理货币，可能导致精度丢失

**建议**: 使用 `github.com/shopspring/decimal` 库

**优先级**: 中等（在处理大额交易时需要修复）

---

### 2. 缺少速率限制 ⚠️ MEDIUM

**问题**: 没有实现速率限制，容易受到暴力破解和API滥用

**建议**: 实现Token Bucket或Leaky Bucket算法

**优先级**: 中等（生产环境建议添加）

---

## 部署注意事项

### 必须设置的环境变量

```bash
# 必需 - OpenAI API密钥
OPENAI_API_KEY=sk-your-key-here

# 必需 - JWT密钥（至少32字符）
JWT_SECRET=your-very-long-and-secure-secret-key-here-min-32-chars

# 必需 - 数据库密码
DB_PASSWORD=your-secure-db-password
```

### 生产环境建议

1. **HTTPS**: 必须启用SSL/TLS
2. **JWT存储**: 考虑使用HttpOnly cookies替代localStorage
3. **CORS**: 更新AllowOrigins为实际域名
4. **速率限制**: 添加API速率限制
5. **监控**: 设置日志监控和告警
6. **备份**: 定期备份数据库

---

## 测试建议

### 安全测试

1. **JWT测试**:
   ```bash
   # 测试无效的user_id格式
   curl -H "Authorization: Bearer <token-with-invalid-uuid>" \
        http://localhost:8080/api/auth/me
   ```

2. **余额测试**:
   ```bash
   # 测试零余额用户调用API
   curl -X POST http://localhost:8080/v1/chat/completions \
        -H "Authorization: Bearer <api-key>" \
        -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"test"}]}'
   ```

3. **JWT密钥测试**:
   ```bash
   # 测试未设置JWT_SECRET
   unset JWT_SECRET
   go run cmd/gateway/main.go
   # 应该输出: JWT_SECRET is required
   ```

---

## 提交信息

**Commit**: bfa022b
**分支**: main
**推送时间**: 2026-01-19

**修改文件**:
- `.gitignore`
- `cmd/gateway/main.go`
- `frontend/src/lib/stores/auth.ts`
- `go.mod`
- `go.sum`
- `internal/config/config.go`
- `internal/handlers/proxy.go`
- `internal/middleware/jwt.go`

---

## 审计工具

本次安全修复基于以下工具的审计结果：
- **Gemini Code Audit**: 全面的代码安全审计
- **Claude Code Review**: 代码质量和最佳实践审查

---

## 联系方式

如有安全问题，请通过以下方式报告：
- GitHub Issues: https://github.com/1307929582/codex/issues
- 标记为 `security` 标签

---

**最后更新**: 2026-01-19
**版本**: v1.0.1 (Security Patch)
