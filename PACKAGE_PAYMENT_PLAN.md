# 套餐+支付系统实现计划

## 功能概述

实现基于Linux Do Credit的套餐购买和每日限额管理系统。

## 核心功能

### 1. 套餐管理（管理员）
- 创建/编辑/删除套餐
- 配置套餐参数：名称、价格、有效期、每日限额
- 套餐排序和状态管理

### 2. Credit支付配置（管理员）
- 配置Credit PID、Key
- 配置回调URL和返回URL
- 启用/禁用Credit支付

### 3. 套餐购买（用户）
- 浏览可用套餐
- 选择套餐并跳转到Credit支付
- 支付成功后自动激活套餐

### 4. 每日限额管理
- 用户有活跃套餐时，优先使用套餐额度
- 每日0点（UTC+8）重置当日使用量
- 套餐额度用完后，从余额扣费
- 套餐到期后自动切换回余额模式

### 5. 支付回调处理
- 接收Credit异步通知
- 验证签名
- 创建用户套餐记录
- 返回success

## 数据库表

1. **packages** - 套餐配置
2. **user_packages** - 用户购买的套餐
3. **daily_usage** - 每日使用记录
4. **payment_orders** - 支付订单
5. **system_settings** - 添加Credit配置字段

## API端点

### 管理员API
- `GET /api/admin/packages` - 获取套餐列表
- `POST /api/admin/packages` - 创建套餐
- `PUT /api/admin/packages/:id` - 更新套餐
- `DELETE /api/admin/packages/:id` - 删除套餐
- `PUT /api/admin/packages/:id/status` - 更新套餐状态

### 用户API
- `GET /api/packages` - 获取可用套餐列表
- `POST /api/packages/:id/purchase` - 购买套餐（创建支付订单）
- `GET /api/user/packages` - 获取我的套餐
- `GET /api/user/daily-usage` - 获取每日使用情况

### 支付回调
- `GET /api/payment/credit/notify` - Credit异步通知
- `GET /api/payment/credit/return` - Credit同步返回

## 计费逻辑

```go
func CalculateCost(userID, cost) {
    // 1. 检查用户是否有活跃套餐
    activePackage := GetActivePackage(userID)

    if activePackage != nil {
        // 2. 获取今日使用记录
        today := time.Now().In(AsiaShanghai).Format("2006-01-02")
        dailyUsage := GetOrCreateDailyUsage(userID, today)

        // 3. 计算剩余额度
        remaining := activePackage.DailyLimit - dailyUsage.UsedAmount

        if remaining >= cost {
            // 套餐额度足够，从套餐扣除
            dailyUsage.UsedAmount += cost
            UpdateDailyUsage(dailyUsage)
            return
        } else if remaining > 0 {
            // 套餐额度不足，部分从套餐扣除，部分从余额扣除
            dailyUsage.UsedAmount += remaining
            UpdateDailyUsage(dailyUsage)
            cost -= remaining
        }
    }

    // 4. 从余额扣除
    DeductBalance(userID, cost)
}
```

## 前端页面

### 管理员
- `/admin/packages` - 套餐管理页面
- `/admin/settings` - 添加Credit配置区域

### 用户
- `/packages` - 套餐购买页面
- `/dashboard` - 显示当前套餐状态和每日使用情况

## 实现步骤

1. ✅ 数据库迁移脚本
2. ✅ Go模型定义
3. 后端API实现
4. Credit支付集成
5. 计费逻辑修改
6. 前端页面开发
7. 测试和部署

## 注意事项

1. **时区处理**：所有日期相关操作使用UTC+8（Asia/Shanghai）
2. **签名验证**：严格按照Credit文档实现MD5签名
3. **幂等性**：支付回调需要处理重复通知
4. **事务处理**：套餐激活和余额变动需要在事务中完成
5. **套餐优先级**：用户可能同时有多个套餐，需要按到期时间排序
