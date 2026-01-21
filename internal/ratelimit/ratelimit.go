package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"
)

type Config struct {
	Enabled           bool
	RequestsPerMinute int
	Burst             int
}

type limiter struct {
	mu       sync.Mutex
	tokens   float64
	last     time.Time
	lastSeen time.Time
}

var (
	configValue atomic.Value
	limiters    sync.Map
	cleanupOnce sync.Once
)

func init() {
	configValue.Store(Config{Enabled: false})
}

func SetConfig(cfg Config) {
	configValue.Store(cfg)
}

func GetConfig() Config {
	value := configValue.Load()
	if value == nil {
		return Config{}
	}
	return value.(Config)
}

func LoadFromDB() {
	var settings models.SystemSettings
	if err := database.DB.First(&settings).Error; err != nil {
		return
	}

	SetConfig(Config{
		Enabled:           settings.RateLimitEnabled,
		RequestsPerMinute: settings.RateLimitRPM,
		Burst:             settings.RateLimitBurst,
	})
}

func Allow(key string) bool {
	cfg := GetConfig()
	if !cfg.Enabled || cfg.RequestsPerMinute <= 0 {
		return true
	}
	if key == "" {
		return true
	}

	startCleanup()

	capacity := float64(cfg.Burst)
	if capacity <= 0 {
		capacity = float64(cfg.RequestsPerMinute)
	}
	rate := float64(cfg.RequestsPerMinute) / 60.0
	now := time.Now()

	value, _ := limiters.LoadOrStore(key, &limiter{
		tokens:   capacity,
		last:     now,
		lastSeen: now,
	})
	bucket := value.(*limiter)

	bucket.mu.Lock()
	elapsed := now.Sub(bucket.last).Seconds()
	if elapsed > 0 {
		bucket.tokens += elapsed * rate
		if bucket.tokens > capacity {
			bucket.tokens = capacity
		}
	}

	allowed := bucket.tokens >= 1
	if allowed {
		bucket.tokens -= 1
	}
	bucket.last = now
	bucket.lastSeen = now
	bucket.mu.Unlock()

	return allowed
}

func startCleanup() {
	cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(10 * time.Minute)
			for range ticker.C {
				cutoff := time.Now().Add(-1 * time.Hour)
				limiters.Range(func(key, value any) bool {
					bucket := value.(*limiter)
					bucket.mu.Lock()
					lastSeen := bucket.lastSeen
					bucket.mu.Unlock()
					if lastSeen.Before(cutoff) {
						limiters.Delete(key)
					}
					return true
				})
			}
		}()
	})
}
