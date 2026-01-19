package upstream

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/google/uuid"
)

// UpstreamSelector manages Codex upstream selection with user affinity
type UpstreamSelector struct {
	mu              sync.RWMutex
	upstreams       []models.CodexUpstream
	lastRefresh     time.Time
	refreshInterval time.Duration
}

var (
	selector *UpstreamSelector
	once     sync.Once
)

// GetSelector returns the singleton upstream selector
func GetSelector() *UpstreamSelector {
	once.Do(func() {
		selector = &UpstreamSelector{
			upstreams:       make([]models.CodexUpstream, 0),
			refreshInterval: 30 * time.Second,
		}
		selector.RefreshUpstreams()
	})
	return selector
}

// RefreshUpstreams reloads upstream configurations from database
func (s *UpstreamSelector) RefreshUpstreams() error {
	var upstreams []models.CodexUpstream

	// Load active upstreams ordered by priority
	if err := database.DB.Where("status IN ?", []string{"active", "unhealthy"}).
		Order("priority ASC, id ASC").
		Find(&upstreams).Error; err != nil {
		return fmt.Errorf("failed to load upstreams: %w", err)
	}

	s.mu.Lock()
	s.upstreams = upstreams
	s.lastRefresh = time.Now()
	s.mu.Unlock()

	log.Printf("[Upstream] Loaded %d upstreams", len(upstreams))
	return nil
}

// SelectForUser selects an upstream for a specific user using consistent hashing
// This ensures the same user always gets the same upstream (session affinity)
func (s *UpstreamSelector) SelectForUser(userID uuid.UUID) (*models.CodexUpstream, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Auto-refresh if needed
	if time.Since(s.lastRefresh) > s.refreshInterval {
		go s.RefreshUpstreams()
	}

	// Filter active upstreams
	activeUpstreams := make([]models.CodexUpstream, 0)
	for _, upstream := range s.upstreams {
		if upstream.Status == "active" {
			activeUpstreams = append(activeUpstreams, upstream)
		}
	}

	if len(activeUpstreams) == 0 {
		return nil, fmt.Errorf("no active upstreams available")
	}

	// Use consistent hashing to select upstream based on user ID
	hash := hashUserID(userID)
	index := int(hash % uint64(len(activeUpstreams)))

	selected := &activeUpstreams[index]
	log.Printf("[Upstream] User %s → Upstream %s (%s)", userID, selected.Name, selected.BaseURL)

	return selected, nil
}

// SelectWithFallback selects an upstream with fallback to next priority
func (s *UpstreamSelector) SelectWithFallback(userID uuid.UUID, excludeIDs []uint) (*models.CodexUpstream, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter active upstreams excluding failed ones
	availableUpstreams := make([]models.CodexUpstream, 0)
	for _, upstream := range s.upstreams {
		if upstream.Status == "active" && !contains(excludeIDs, upstream.ID) {
			availableUpstreams = append(availableUpstreams, upstream)
		}
	}

	if len(availableUpstreams) == 0 {
		return nil, fmt.Errorf("no available upstreams (all failed or disabled)")
	}

	// Use consistent hashing
	hash := hashUserID(userID)
	index := int(hash % uint64(len(availableUpstreams)))

	selected := &availableUpstreams[index]
	log.Printf("[Upstream] User %s → Fallback Upstream %s (%s)", userID, selected.Name, selected.BaseURL)

	return selected, nil
}

// GetAllUpstreams returns all upstreams (for admin management)
func (s *UpstreamSelector) GetAllUpstreams() []models.CodexUpstream {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.CodexUpstream, len(s.upstreams))
	copy(result, s.upstreams)
	return result
}

// hashUserID creates a consistent hash from user ID
func hashUserID(userID uuid.UUID) uint64 {
	h := sha256.Sum256(userID[:])
	return binary.BigEndian.Uint64(h[:8])
}

// contains checks if a slice contains a value
func contains(slice []uint, val uint) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
