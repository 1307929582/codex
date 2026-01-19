package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Balance      float64        `gorm:"type:decimal(18,6);default:0" json:"balance"`
	Status       string         `gorm:"type:varchar(20);default:'active'" json:"status"`
	Role         string         `gorm:"type:varchar(20);default:'user'" json:"role"` // user, admin, super_admin
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
	ID                 uint      `gorm:"primaryKey" json:"id"`
	ModelName          string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"model_name"`
	InputPricePer1k    float64   `gorm:"type:decimal(10,6);not null" json:"input_price_per_1k"`
	OutputPricePer1k   float64   `gorm:"type:decimal(10,6);not null" json:"output_price_per_1k"`
	MarkupMultiplier   float64   `gorm:"type:decimal(4,2);default:1.5" json:"markup_multiplier"`
	EffectiveFrom      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"effective_from"`
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
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
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
