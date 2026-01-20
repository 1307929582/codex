package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255)" json:"-"` // Optional for OAuth users
	Balance      float64        `gorm:"type:decimal(18,6);default:0" json:"balance"`
	Status       string         `gorm:"type:varchar(20);default:'active'" json:"status"`
	Role         string         `gorm:"type:varchar(20);default:'user'" json:"role"` // user, admin, super_admin

	// OAuth fields
	OAuthProvider string `gorm:"type:varchar(50)" json:"oauth_provider"` // "linuxdo", "email", etc.
	OAuthID       string `gorm:"type:varchar(255);index:idx_oauth" json:"oauth_id"` // Provider's user ID
	Username      string `gorm:"type:varchar(100)" json:"username"` // Display name from OAuth
	AvatarURL     string `gorm:"type:varchar(500)" json:"avatar_url"` // Profile picture URL

	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type APIKey struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_id" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"-"`
	KeyHash     string         `gorm:"type:varchar(64);not null;uniqueIndex:idx_key_hash" json:"-"`
	KeyPrefix   string         `gorm:"type:varchar(16);not null" json:"key_prefix"`
	Name        string         `gorm:"type:varchar(100)" json:"name"`
	QuotaLimit  *float64       `gorm:"type:decimal(18,6)" json:"quota_limit"`
	TotalUsage  int64          `gorm:"default:0" json:"total_usage"`
	Status      string         `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	LastUsedAt  *time.Time     `json:"last_used_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type ModelPricing struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	ModelName            string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"model_name"`
	InputPricePer1k      float64   `gorm:"type:decimal(10,6);not null" json:"input_price_per_1k"`
	OutputPricePer1k     float64   `gorm:"type:decimal(10,6);not null" json:"output_price_per_1k"`
	CacheReadPricePer1k  float64   `gorm:"type:decimal(10,6);default:0" json:"cache_read_price_per_1k"` // Cache read tokens pricing (usually 10% of input price)
	MarkupMultiplier     float64   `gorm:"type:decimal(4,2);default:1.5" json:"markup_multiplier"`
	EffectiveFrom        time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"effective_from"`
}

type UsageLog struct {
	RequestID    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"request_id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_user_created" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"-"`
	APIKeyID     uint      `gorm:"not null;index:idx_api_key_created" json:"api_key_id"`
	APIKey       APIKey    `gorm:"foreignKey:APIKeyID" json:"-"`
	Model        string    `gorm:"type:varchar(100)" json:"model"`
	InputTokens  int       `gorm:"not null" json:"input_tokens"`
	OutputTokens int       `gorm:"not null" json:"output_tokens"`
	CachedTokens int       `gorm:"default:0" json:"cached_tokens"` // Cached input tokens
	TotalTokens  int       `gorm:"not null" json:"total_tokens"`
	Cost         float64   `gorm:"type:decimal(18,6);not null" json:"cost"`
	LatencyMs    int       `json:"latency_ms"`
	StatusCode   int       `json:"status_code"`
	CreatedAt    time.Time `gorm:"index:idx_user_created,idx_api_key_created" json:"created_at"`
}

func (UsageLog) TableName() string {
	return "usage_logs"
}

type Transaction struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index:idx_txn_user" json:"user_id"`
	Amount      float64   `gorm:"type:decimal(18,6);not null" json:"amount"`
	Type        string    `gorm:"type:varchar(20);not null" json:"type"` // deposit, refund
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

type SystemSettings struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Announcement        string    `gorm:"type:text" json:"announcement"`
	DefaultBalance      float64   `gorm:"type:decimal(18,6);default:0" json:"default_balance"`
	MinRechargeAmount   float64   `gorm:"type:decimal(18,6);default:10" json:"min_recharge_amount"`
	RegistrationEnabled bool      `gorm:"default:true" json:"registration_enabled"`
	OpenAIAPIKey        string    `gorm:"type:varchar(255)" json:"openai_api_key"`
	OpenAIBaseURL       string    `gorm:"type:varchar(255);default:'https://api.openai.com/v1'" json:"openai_base_url"`

	// LinuxDo OAuth Settings
	LinuxDoClientID     string `gorm:"column:linuxdo_client_id;type:varchar(255)" json:"linuxdo_client_id"`
	LinuxDoClientSecret string `gorm:"column:linuxdo_client_secret;type:varchar(255)" json:"linuxdo_client_secret"`
	LinuxDoEnabled      bool   `gorm:"column:linuxdo_enabled;default:false" json:"linuxdo_enabled"`

	// Credit Payment Settings
	CreditEnabled   bool   `gorm:"column:credit_enabled;default:false" json:"credit_enabled"`
	CreditPID       string `gorm:"column:credit_pid;type:varchar(255)" json:"credit_pid"`
	CreditKey       string `gorm:"column:credit_key;type:varchar(255)" json:"credit_key"`
	CreditNotifyURL string `gorm:"column:credit_notify_url;type:varchar(500)" json:"credit_notify_url"`
	CreditReturnURL string `gorm:"column:credit_return_url;type:varchar(500)" json:"credit_return_url"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Package struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description"`
	Price        float64   `gorm:"type:decimal(18,6);not null" json:"price"`
	DurationDays int       `gorm:"not null" json:"duration_days"`
	DailyLimit   float64   `gorm:"type:decimal(18,6);not null" json:"daily_limit"`
	Status       string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	SortOrder    int       `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserPackage struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"-"`
	PackageID    uint      `gorm:"not null" json:"package_id"`
	Package      Package   `gorm:"foreignKey:PackageID" json:"-"`
	PackageName  string    `gorm:"type:varchar(100);not null" json:"package_name"`
	PackagePrice float64   `gorm:"type:decimal(18,6);not null" json:"package_price"`
	DurationDays int       `gorm:"not null" json:"duration_days"`
	DailyLimit   float64   `gorm:"type:decimal(18,6);not null" json:"daily_limit"`
	StartDate    time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate      time.Time `gorm:"type:date;not null" json:"end_date"`
	Status       string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DailyUsage struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User          User       `gorm:"foreignKey:UserID" json:"-"`
	UserPackageID *uuid.UUID `gorm:"type:uuid" json:"user_package_id"`
	Date          time.Time  `gorm:"type:date;not null" json:"date"`
	UsedAmount    float64    `gorm:"type:decimal(18,6);default:0" json:"used_amount"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (DailyUsage) TableName() string {
	return "daily_usage"
}

type PaymentOrder struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User          User       `gorm:"foreignKey:UserID" json:"-"`
	PackageID     *uint      `json:"package_id"`
	Package       *Package   `gorm:"foreignKey:PackageID" json:"-"`
	OrderNo       string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"`
	OutTradeNo    string     `gorm:"type:varchar(64)" json:"out_trade_no"`
	TradeNo       string     `gorm:"type:varchar(64)" json:"trade_no"`
	Amount        float64    `gorm:"type:decimal(18,6);not null" json:"amount"`
	Status        string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	PaymentMethod string     `gorm:"type:varchar(50);default:'credit'" json:"payment_method"`
	PaymentData   string     `gorm:"type:text" json:"payment_data"`
	NotifyData    string     `gorm:"type:text" json:"notify_data"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	PaidAt        *time.Time `json:"paid_at"`
}

type AdminLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AdminID   uuid.UUID `gorm:"type:uuid;not null;index:idx_admin_logs" json:"admin_id"`
	Admin     User      `gorm:"foreignKey:AdminID" json:"-"`
	Action    string    `gorm:"type:varchar(100);not null" json:"action"`
	Target    string    `gorm:"type:varchar(100)" json:"target"`
	Details   string    `gorm:"type:text" json:"details"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
	CreatedAt time.Time `gorm:"index:idx_admin_logs" json:"created_at"`
}

type CodexUpstream struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	BaseURL     string    `gorm:"type:varchar(255);not null" json:"base_url"`
	APIKey      string    `gorm:"type:varchar(255);not null" json:"api_key"`
	Priority    int       `gorm:"default:0;index" json:"priority"` // Lower number = higher priority
	Status      string    `gorm:"type:varchar(20);default:'active'" json:"status"` // active, disabled, unhealthy
	Weight      int       `gorm:"default:1" json:"weight"` // For load balancing (not used with user affinity)
	MaxRetries  int       `gorm:"default:3" json:"max_retries"`
	Timeout     int       `gorm:"default:120" json:"timeout"` // Seconds
	HealthCheck string    `gorm:"type:varchar(255)" json:"health_check"` // Health check endpoint
	LastChecked *time.Time `json:"last_checked"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
