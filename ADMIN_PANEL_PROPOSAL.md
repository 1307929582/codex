# 管理员面板设计方案

## 📋 功能规划

### 1. 系统配置管理
- ✅ **模型定价管理**（已实现，在数据库中）
  - 添加/编辑/删除模型
  - 设置输入/输出价格
  - 设置加价倍数

- 🆕 **系统设置**
  - 系统公告（显示在Dashboard）
  - 新用户默认余额
  - 最小充值金额
  - API调用速率限制
  - 功能开关（注册开关、充值开关等）

### 2. 用户管理
- 🆕 **用户列表**
  - 查看所有用户
  - 搜索/筛选用户
  - 查看用户详情（余额、使用量、API密钥）

- 🆕 **用户操作**
  - 手动调整用户余额
  - 禁用/启用用户
  - 重置用户密码
  - 查看用户使用记录

### 3. 财务管理
- 🆕 **交易记录**
  - 查看所有充值记录
  - 查看所有消费记录
  - 导出财务报表

- 🆕 **统计分析**
  - 总收入/总支出
  - 用户增长趋势
  - API调用统计
  - 热门模型排行

### 4. 监控告警
- 🆕 **系统监控**
  - 实时API调用量
  - 错误率监控
  - 响应时间统计

- 🆕 **告警设置**
  - 余额不足告警
  - 异常调用告警
  - 系统错误告警

### 5. 日志审计
- 🆕 **操作日志**
  - 管理员操作记录
  - 用户登录记录
  - 敏感操作审计

---

## 🎨 界面设计

### 路由结构
```
/admin
├── /dashboard          # 管理员首页（统计概览）
├── /users              # 用户管理
│   ├── /list          # 用户列表
│   └── /[id]          # 用户详情
├── /pricing            # 模型定价管理
├── /transactions       # 交易记录
├── /settings           # 系统设置
├── /monitoring         # 系统监控
└── /logs              # 操作日志
```

### 权限控制
```typescript
// 添加管理员角色
enum UserRole {
  USER = 'user',
  ADMIN = 'admin',
  SUPER_ADMIN = 'super_admin'
}

// 在User模型中添加role字段
type User = {
  id: string;
  email: string;
  role: UserRole;  // 新增
  balance: number;
  // ...
}
```

---

## 🔧 技术实现

### 后端API（Go）

#### 1. 新增管理员中间件
```go
// internal/middleware/admin.go
func AdminAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("user").(models.User)

        if user.Role != "admin" && user.Role != "super_admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

#### 2. 新增管理员API端点
```go
// cmd/gateway/main.go
admin := apiGroup.Group("/admin")
admin.Use(middleware.JWTAuthMiddleware())
admin.Use(middleware.AdminAuthMiddleware())
{
    // 用户管理
    admin.GET("/users", handlers.AdminListUsers)
    admin.GET("/users/:id", handlers.AdminGetUser)
    admin.PUT("/users/:id/balance", handlers.AdminUpdateBalance)
    admin.PUT("/users/:id/status", handlers.AdminUpdateUserStatus)

    // 系统设置
    admin.GET("/settings", handlers.AdminGetSettings)
    admin.PUT("/settings", handlers.AdminUpdateSettings)

    // 统计分析
    admin.GET("/stats/overview", handlers.AdminGetOverview)
    admin.GET("/stats/revenue", handlers.AdminGetRevenue)

    // 交易记录
    admin.GET("/transactions", handlers.AdminGetTransactions)

    // 操作日志
    admin.GET("/logs", handlers.AdminGetLogs)
}
```

#### 3. 新增数据模型
```go
// internal/models/models.go

// 系统设置
type SystemSettings struct {
    ID                  uint    `gorm:"primaryKey"`
    Announcement        string  `gorm:"type:text"`
    DefaultBalance      float64 `gorm:"default:0"`
    MinRechargeAmount   float64 `gorm:"default:10"`
    RegistrationEnabled bool    `gorm:"default:true"`
    CreatedAt           time.Time
    UpdatedAt           time.Time
}

// 操作日志
type AdminLog struct {
    ID        uint      `gorm:"primaryKey"`
    AdminID   uuid.UUID `gorm:"type:uuid;not null"`
    Action    string    `gorm:"type:varchar(100);not null"`
    Target    string    `gorm:"type:varchar(100)"`
    Details   string    `gorm:"type:text"`
    IPAddress string    `gorm:"type:varchar(45)"`
    CreatedAt time.Time
}
```

### 前端页面（Next.js）

#### 1. 管理员布局
```typescript
// frontend/src/app/admin/layout.tsx
export default function AdminLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="flex h-screen">
      {/* 侧边栏 */}
      <AdminSidebar />

      {/* 主内容区 */}
      <main className="flex-1 overflow-y-auto p-8">
        {children}
      </main>
    </div>
  )
}
```

#### 2. 用户管理页面
```typescript
// frontend/src/app/admin/users/page.tsx
export default function AdminUsersPage() {
  const { data: users } = useQuery({
    queryKey: ['admin', 'users'],
    queryFn: () => api.get('/api/admin/users'),
  })

  return (
    <div>
      <h1>用户管理</h1>

      {/* 搜索和筛选 */}
      <UserFilters />

      {/* 用户列表表格 */}
      <UsersTable users={users} />
    </div>
  )
}
```

---

## 🚀 实施步骤

### Phase 1: 基础功能（2-3小时）
1. ✅ 添加User.role字段和数据库迁移
2. ✅ 实现管理员中间件
3. ✅ 创建管理员布局和侧边栏
4. ✅ 实现用户列表和详情页面

### Phase 2: 核心功能（3-4小时）
1. ✅ 实现用户余额调整功能
2. ✅ 实现系统设置管理
3. ✅ 实现交易记录查看
4. ✅ 添加统计分析Dashboard

### Phase 3: 高级功能（2-3小时）
1. ✅ 实现操作日志记录
2. ✅ 添加实时监控
3. ✅ 实现数据导出功能
4. ✅ 添加告警通知

---

## 🔒 安全考虑

### 1. 权限验证
- 所有管理员API必须经过双重验证（JWT + Admin Role）
- 敏感操作需要二次确认
- 记录所有管理员操作日志

### 2. 审计追踪
- 记录管理员的所有操作
- 包含IP地址、时间戳、操作详情
- 不可删除的审计日志

### 3. 初始管理员
```bash
# 创建初始管理员的方式
# 方式1: 通过环境变量
INITIAL_ADMIN_EMAIL=admin@example.com
INITIAL_ADMIN_PASSWORD=change-me-immediately

# 方式2: 通过CLI命令
go run cmd/gateway/main.go create-admin --email admin@example.com
```

---

## 💡 环境变量 vs 管理员面板

### 保留环境变量的配置（不应该在面板中）
```bash
# 基础设施级 - 安全敏感
OPENAI_API_KEY=sk-xxx
JWT_SECRET=xxx
DB_PASSWORD=xxx
DB_HOST=localhost
DB_PORT=5432

# 初始化配置 - 只在首次启动时使用
INITIAL_ADMIN_EMAIL=admin@example.com
INITIAL_ADMIN_PASSWORD=xxx
```

### 移到管理员面板的配置（业务级）
- ✅ 模型定价
- ✅ 新用户默认余额
- ✅ 系统公告
- ✅ 功能开关
- ✅ 速率限制配置

---

## 📊 预期效果

### 使用体验改进
**之前**:
1. 修改模型定价 → 编辑数据库 → 重启服务
2. 调整用户余额 → 写SQL语句 → 手动执行
3. 查看统计数据 → 写复杂查询 → 导出Excel

**之后**:
1. 修改模型定价 → 打开管理面板 → 点击编辑 → 保存
2. 调整用户余额 → 搜索用户 → 点击调整 → 输入金额
3. 查看统计数据 → 打开Dashboard → 实时图表展示

---

## 🎯 是否需要实现？

请告诉我：
1. **是否需要管理员面板？**
2. **优先实现哪些功能？**（用户管理、统计分析、系统设置等）
3. **是否需要保留环境变量配置？**（推荐保留用于敏感信息）

我可以立即开始实现！
