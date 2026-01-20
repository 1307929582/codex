# 安全修复和测试指南

## 🔒 本次修复的关键问题

### 1. 并发安全问题（Critical）

#### 问题：余额扣费存在竞态条件
**修复前**：
```go
if user.Balance < cost {
    return fmt.Errorf("insufficient balance")
}
user.Balance -= cost
tx.Save(&user)
```
**问题**：两个并发请求可能同时通过余额检查，导致余额为负

**修复后**：
```go
result := tx.Exec("UPDATE users SET balance = balance - ? WHERE id = ? AND balance >= ?", cost, userID, cost)
if result.RowsAffected == 0 {
    return fmt.Errorf("insufficient balance")
}
```
**效果**：使用数据库原子操作，确保余额永远不会为负

#### 问题：每日使用量更新不是原子操作
**修复前**：
```go
dailyUsage.UsedAmount += cost
tx.Save(&dailyUsage)
```
**问题**：并发请求可能导致使用量统计错误

**修复后**：
```go
result := tx.Model(&models.DailyUsage{}).
    Where("id = ? AND used_amount + ? <= ?", dailyUsage.ID, cost, activePackage.DailyLimit).
    Update("used_amount", gorm.Expr("used_amount + ?", cost))
if result.RowsAffected == 0 {
    return fmt.Errorf("concurrent update conflict or quota exceeded")
}
```
**效果**：原子更新，同时检查配额限制

### 2. 支付安全问题（Critical）

#### 增强的安全验证
**新增检查**：
1. ✅ 订单过期检查（24小时）- 防止重放攻击
2. ✅ trade_no 非空验证
3. ✅ 详细的安全日志记录
4. ✅ IP地址记录

**修复后的代码**：
```go
// Check if order is too old (prevent replay attacks)
if time.Since(order.CreatedAt) > 24*time.Hour {
    log.Printf("[Payment] Order too old: %s, created at: %s", outTradeNo, order.CreatedAt)
    c.String(http.StatusBadRequest, "order expired")
    return
}

// Validate trade_no is not empty
if tradeNo == "" {
    log.Printf("[Payment] Empty trade_no from IP: %s", c.ClientIP())
    c.String(http.StatusBadRequest, "invalid trade_no")
    return
}
```

### 3. 网络连接问题（High）

#### 问题：前端无法解析backend主机名
**错误信息**：
```
Error: getaddrinfo ENOTFOUND backend
```

**原因**：Docker容器没有在同一个网络中

**修复**：
```yaml
networks:
  codex-network:
    driver: bridge

services:
  backend:
    networks:
      - codex-network
  frontend:
    networks:
      - codex-network
    environment:
      INTERNAL_API_URL: http://backend:12322
```

## 🧪 测试指南

### 测试1：并发安全测试

#### 测试余额扣费并发安全
```bash
# 创建测试脚本
cat > test_concurrent_billing.sh << 'EOF'
#!/bin/bash
USER_TOKEN="your_user_token"
API_URL="http://localhost:12322"

# 并发发送10个请求
for i in {1..10}; do
  curl -X POST "$API_URL/v1/chat/completions" \
    -H "Authorization: Bearer $USER_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "model": "gpt-4",
      "messages": [{"role": "user", "content": "test"}]
    }' &
done
wait
EOF

chmod +x test_concurrent_billing.sh
./test_concurrent_billing.sh
```

**预期结果**：
- 所有请求都应该成功或失败（不会出现余额为负）
- 检查数据库：`SELECT balance FROM users WHERE id = 'user_id'`
- 余额应该是正确的值

#### 测试每日配额并发安全
```bash
# 假设用户有套餐，每日限额$5
# 并发发送多个请求，总费用超过$5

# 检查每日使用量
docker exec codex-postgres psql -U postgres -d codex_gateway -c "
SELECT user_id, date, used_amount, user_package_id
FROM daily_usage
WHERE date = CURRENT_DATE;
"
```

**预期结果**：
- used_amount 不应该超过 daily_limit
- 超出部分应该从余额扣除

### 测试2：支付安全测试

#### 测试订单过期保护
```bash
# 1. 创建一个测试订单
# 2. 修改数据库中的created_at为25小时前
docker exec codex-postgres psql -U postgres -d codex_gateway -c "
UPDATE payment_orders
SET created_at = NOW() - INTERVAL '25 hours'
WHERE order_no = 'TEST_ORDER_NO';
"

# 3. 尝试回调
curl -X GET "http://localhost:12322/api/payment/credit/notify?..."
```

**预期结果**：
- 返回 "order expired"
- 订单状态不变

#### 测试重放攻击保护
```bash
# 1. 完成一次正常支付
# 2. 重复发送相同的回调请求
curl -X GET "http://localhost:12322/api/payment/credit/notify?..."
```

**预期结果**：
- 第一次：创建套餐，返回 "success"
- 第二次：返回 "success"，但不创建重复套餐
- 检查日志：应该看到 "Order already paid" 日志

### 测试3：网络连接测试

#### 测试前端到后端的连接
```bash
# 1. 检查网络
docker network inspect codex_codex-network

# 2. 测试前端容器内的DNS解析
docker exec codex-frontend ping -c 3 backend

# 3. 测试前端容器内的HTTP连接
docker exec codex-frontend curl -s http://backend:12322/health
```

**预期结果**：
- ping 成功
- curl 返回 `{"status":"ok"}`

#### 测试前端页面加载
```bash
# 访问前端页面
curl -s http://localhost:12321 | grep -i "codex"
```

**预期结果**：
- 返回HTML内容
- 不应该有 "ENOTFOUND backend" 错误

### 测试4：完整购买流程测试

#### 步骤1：配置Credit支付
1. 登录管理员账号
2. 进入系统设置
3. 配置Credit参数
4. 启用Credit支付

#### 步骤2：创建测试套餐
1. 进入套餐管理
2. 创建测试套餐：
   - 名称：测试套餐
   - 价格：0.01
   - 有效期：1天
   - 每日限额：0.01

#### 步骤3：购买套餐
1. 用户登录
2. 访问套餐页面
3. 点击购买
4. 完成支付（使用测试环境）

#### 步骤4：验证套餐激活
```bash
# 检查用户套餐
docker exec codex-postgres psql -U postgres -d codex_gateway -c "
SELECT id, user_id, package_name, daily_limit, start_date, end_date, status
FROM user_packages
WHERE status = 'active';
"

# 检查支付订单
docker exec codex-postgres psql -U postgres -d codex_gateway -c "
SELECT order_no, amount, status, paid_at
FROM payment_orders
WHERE status = 'paid'
ORDER BY created_at DESC
LIMIT 5;
"
```

#### 步骤5：测试每日限额
1. 发起API请求
2. 检查Dashboard显示的使用情况
3. 验证额度正确扣除

## 📊 监控和日志

### 查看支付日志
```bash
docker compose logs backend | grep "\[Payment\]"
```

### 查看计费日志
```bash
docker compose logs backend | grep -i "billing\|balance\|quota"
```

### 查看错误日志
```bash
docker compose logs backend | grep -i "error\|failed"
```

### 实时监控
```bash
# 监控所有日志
docker compose logs -f

# 只监控后端
docker compose logs -f backend

# 只监控前端
docker compose logs -f frontend
```

## 🔍 故障排查

### 问题1：前端仍然无法连接后端
```bash
# 检查网络
docker network ls
docker network inspect codex_codex-network

# 检查容器是否在同一网络
docker inspect codex-frontend | grep -A 10 Networks
docker inspect codex-backend | grep -A 10 Networks

# 重建网络
docker compose down
docker network prune
docker compose up -d
```

### 问题2：余额变为负数
```bash
# 检查余额
docker exec codex-postgres psql -U postgres -d codex_gateway -c "
SELECT id, email, balance FROM users WHERE balance < 0;
"

# 如果发现负余额，说明修复未生效
# 检查代码版本
git log --oneline -5
```

### 问题3：支付回调失败
```bash
# 检查订单状态
docker exec codex-postgres psql -U postgres -d codex_gateway -c "
SELECT order_no, status, created_at, paid_at, notify_data
FROM payment_orders
ORDER BY created_at DESC
LIMIT 10;
"

# 检查支付日志
docker compose logs backend | grep "\[Payment\]" | tail -50
```

## ✅ 验收标准

### 并发安全
- [ ] 10个并发请求后余额正确
- [ ] 每日使用量不超过限额
- [ ] 没有数据库死锁错误

### 支付安全
- [ ] 过期订单无法回调
- [ ] 重复回调不创建重复套餐
- [ ] 所有支付事件都有日志

### 网络连接
- [ ] 前端可以访问
- [ ] 前端可以调用后端API
- [ ] 没有DNS解析错误

### 功能完整性
- [ ] 可以创建套餐
- [ ] 可以购买套餐
- [ ] 套餐正确激活
- [ ] 每日限额正确扣除
- [ ] Dashboard正确显示

## 🚀 部署到生产环境

### 部署前检查
1. [ ] 所有测试通过
2. [ ] 数据库已备份
3. [ ] Credit配置已准备
4. [ ] 回调URL可外网访问

### 部署步骤
```bash
# 1. 备份数据库
docker exec codex-postgres pg_dump -U postgres codex_gateway > backup_$(date +%Y%m%d_%H%M%S).sql

# 2. 拉取代码
git pull

# 3. 运行部署脚本
chmod +x deploy-security-fixes.sh
./deploy-security-fixes.sh

# 4. 验证部署
curl http://localhost:12322/health
curl http://localhost:12321

# 5. 监控日志
docker compose logs -f
```

### 回滚计划
如果出现问题：
```bash
# 1. 回滚代码
git reset --hard HEAD~1

# 2. 重新构建
docker compose down
docker compose up -d --build

# 3. 恢复数据库（如果需要）
docker exec -i codex-postgres psql -U postgres codex_gateway < backup_YYYYMMDD_HHMMSS.sql
```

## 📞 支持

如有问题，请检查：
1. 日志文件
2. 数据库状态
3. 网络连接
4. 环境变量配置
