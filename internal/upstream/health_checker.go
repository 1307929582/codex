package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"
)

// HealthChecker manages upstream health checks
type HealthChecker struct {
	mu              sync.RWMutex
	checkInterval   time.Duration
	timeout         time.Duration
	maxFailures     int
	failureCounts   map[uint]int
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

var (
	healthChecker *HealthChecker
	checkerOnce   sync.Once
)

// GetHealthChecker returns the singleton health checker
func GetHealthChecker() *HealthChecker {
	checkerOnce.Do(func() {
		healthChecker = &HealthChecker{
			checkInterval: 60 * time.Second,  // Check every minute
			timeout:       10 * time.Second,  // 10 second timeout
			maxFailures:   3,                 // Mark unhealthy after 3 consecutive failures
			failureCounts: make(map[uint]int),
			stopCh:        make(chan struct{}),
		}
	})
	return healthChecker
}

// Start starts the health checker
func (hc *HealthChecker) Start() {
	hc.wg.Add(1)
	go func() {
		defer hc.wg.Done()
		ticker := time.NewTicker(hc.checkInterval)
		defer ticker.Stop()

		log.Printf("[HealthCheck] Started (interval: %v)", hc.checkInterval)

		// Run initial check
		hc.checkAllUpstreams()

		for {
			select {
			case <-ticker.C:
				hc.checkAllUpstreams()
			case <-hc.stopCh:
				return
			}
		}
	}()
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
	hc.wg.Wait()
	log.Println("[HealthCheck] Stopped")
}

// CheckAllUpstreams checks health of all upstreams (exported for manual trigger)
func (hc *HealthChecker) CheckAllUpstreams() {
	hc.checkAllUpstreams()
}

// checkAllUpstreams checks health of all upstreams
func (hc *HealthChecker) checkAllUpstreams() {
	var upstreams []models.CodexUpstream

	// Load all upstreams (including disabled ones, but we'll only check active/unhealthy)
	if err := database.DB.Find(&upstreams).Error; err != nil {
		log.Printf("[HealthCheck] Failed to load upstreams: %v", err)
		return
	}

	log.Printf("[HealthCheck] Found %d total upstreams", len(upstreams))

	checkedCount := 0
	for _, upstream := range upstreams {
		// Only check active or unhealthy upstreams (skip manually disabled ones)
		if upstream.Status == "active" || upstream.Status == "unhealthy" {
			log.Printf("[HealthCheck] Checking upstream: %s (status: %s, base_url: %s)",
				upstream.Name, upstream.Status, upstream.BaseURL)
			go hc.checkUpstream(&upstream)
			checkedCount++
		} else {
			log.Printf("[HealthCheck] Skipping upstream: %s (status: %s)", upstream.Name, upstream.Status)
		}
	}

	log.Printf("[HealthCheck] Triggered health check for %d/%d upstreams", checkedCount, len(upstreams))
}

// checkUpstream checks a single upstream
func (hc *HealthChecker) checkUpstream(upstream *models.CodexUpstream) {
	healthy := hc.performHealthCheck(upstream)

	hc.mu.Lock()
	defer hc.mu.Unlock()

	if healthy {
		// Reset failure count
		hc.failureCounts[upstream.ID] = 0

		// If was unhealthy, mark as active
		if upstream.Status == "unhealthy" {
			if err := database.DB.Model(&models.CodexUpstream{}).
				Where("id = ?", upstream.ID).
				Updates(map[string]interface{}{
					"status":       "active",
					"last_checked": time.Now(),
				}).Error; err != nil {
				log.Printf("[HealthCheck] Failed to update upstream %s status: %v", upstream.Name, err)
			} else {
				log.Printf("[HealthCheck] ✅ Upstream %s recovered (active)", upstream.Name)
				// Refresh selector
				GetSelector().RefreshUpstreams()
			}
		} else {
			// Just update last_checked
			database.DB.Model(&models.CodexUpstream{}).
				Where("id = ?", upstream.ID).
				Update("last_checked", time.Now())
		}
	} else {
		// Increment failure count
		hc.failureCounts[upstream.ID]++
		failCount := hc.failureCounts[upstream.ID]

		log.Printf("[HealthCheck] ❌ Upstream %s check failed (failures: %d/%d)",
			upstream.Name, failCount, hc.maxFailures)

		// Mark as unhealthy after max failures
		if failCount >= hc.maxFailures && upstream.Status == "active" {
			if err := database.DB.Model(&models.CodexUpstream{}).
				Where("id = ?", upstream.ID).
				Updates(map[string]interface{}{
					"status":       "unhealthy",
					"last_checked": time.Now(),
				}).Error; err != nil {
				log.Printf("[HealthCheck] Failed to update upstream %s status: %v", upstream.Name, err)
			} else {
				log.Printf("[HealthCheck] ⚠️  Upstream %s marked as unhealthy", upstream.Name)
				// Refresh selector
				GetSelector().RefreshUpstreams()
			}
		} else {
			// Just update last_checked
			database.DB.Model(&models.CodexUpstream{}).
				Where("id = ?", upstream.ID).
				Update("last_checked", time.Now())
		}
	}
}

// performHealthCheck performs actual health check
func (hc *HealthChecker) performHealthCheck(upstream *models.CodexUpstream) bool {
	log.Printf("[HealthCheck] Starting health check for %s at %s", upstream.Name, upstream.BaseURL)

	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	// Create a simple test request
	testReq := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "user", "content": "test"},
		},
		"max_tokens": 1,
	}

	reqBytes, err := json.Marshal(testReq)
	if err != nil {
		log.Printf("[HealthCheck] Failed to marshal test request: %v", err)
		return false
	}

	// Create HTTP request
	url := upstream.BaseURL + "/chat/completions"
	log.Printf("[HealthCheck] Sending request to: %s", url)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBytes))
	if err != nil {
		log.Printf("[HealthCheck] Failed to create request for %s: %v", upstream.Name, err)
		return false
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+upstream.APIKey)
	log.Printf("[HealthCheck] Using API Key: %s...", upstream.APIKey[:min(20, len(upstream.APIKey))])

	// Send request
	client := &http.Client{
		Timeout: hc.timeout,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[HealthCheck] Request failed for %s: %v", upstream.Name, err)
		return false
	}
	defer resp.Body.Close()

	// Read response (limit to 1KB)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		log.Printf("[HealthCheck] Failed to read response from %s: %v", upstream.Name, err)
		return false
	}

	log.Printf("[HealthCheck] Response from %s: status=%d, body=%s", upstream.Name, resp.StatusCode, string(body))

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		// 2xx, 3xx, 4xx are considered "healthy" (server is responding)
		// 4xx means auth/request issues, but server is up
		log.Printf("[HealthCheck] ✅ Upstream %s is healthy (status: %d)", upstream.Name, resp.StatusCode)
		return true
	}

	// 5xx errors mean server issues
	log.Printf("[HealthCheck] ❌ Upstream %s returned error (status: %d)", upstream.Name, resp.StatusCode)
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetFailureCount returns the current failure count for an upstream
func (hc *HealthChecker) GetFailureCount(upstreamID uint) int {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.failureCounts[upstreamID]
}
