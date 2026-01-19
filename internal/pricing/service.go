package pricing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"codex-gateway/internal/database"
	"codex-gateway/internal/models"
)

const (
	// LiteLLM pricing URL
	defaultPricingURL = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"

	// Update intervals
	updateInterval = 24 * time.Hour
	checkInterval  = 10 * time.Minute

	// Cache directory
	cacheDir = "./data/pricing"
)

// ModelPricing represents LiteLLM pricing data
type LiteLLMPricing struct {
	InputCostPerToken  float64 `json:"input_cost_per_token"`
	OutputCostPerToken float64 `json:"output_cost_per_token"`
	MaxInputTokens     int     `json:"max_input_tokens"`
	MaxOutputTokens    int     `json:"max_output_tokens"`
}

// PricingService manages automatic pricing updates
type PricingService struct {
	mu          sync.RWMutex
	pricingData map[string]*LiteLLMPricing
	lastUpdated time.Time
	localHash   string
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

var (
	service *PricingService
	once    sync.Once
)

// GetService returns the singleton pricing service
func GetService() *PricingService {
	once.Do(func() {
		service = &PricingService{
			pricingData: make(map[string]*LiteLLMPricing),
			stopCh:      make(chan struct{}),
		}
	})
	return service
}

// Initialize starts the pricing service
func (s *PricingService) Initialize() error {
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("[Pricing] Failed to create cache directory: %v", err)
	}

	// Load initial pricing
	if err := s.loadPricing(); err != nil {
		log.Printf("[Pricing] Initial load failed: %v", err)
		return err
	}

	// Start background updater
	s.startUpdater()

	log.Printf("[Pricing] Service initialized with %d models", len(s.pricingData))
	return nil
}

// Stop stops the pricing service
func (s *PricingService) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	log.Println("[Pricing] Service stopped")
}

// loadPricing loads pricing from cache or downloads it
func (s *PricingService) loadPricing() error {
	cacheFile := filepath.Join(cacheDir, "pricing.json")

	// Check if cache exists and is recent
	if info, err := os.Stat(cacheFile); err == nil {
		age := time.Since(info.ModTime())
		if age < updateInterval {
			// Load from cache
			if err := s.loadFromFile(cacheFile); err == nil {
				log.Printf("[Pricing] Loaded from cache (age: %v)", age.Round(time.Hour))
				return nil
			}
		}
	}

	// Download fresh pricing
	log.Println("[Pricing] Downloading fresh pricing data...")
	return s.downloadPricing()
}

// downloadPricing downloads pricing from LiteLLM
func (s *PricingService) downloadPricing() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", defaultPricingURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	// Parse pricing data
	data, err := s.parsePricing(body)
	if err != nil {
		return fmt.Errorf("parse pricing: %w", err)
	}

	// Save to cache
	cacheFile := filepath.Join(cacheDir, "pricing.json")
	if err := os.WriteFile(cacheFile, body, 0644); err != nil {
		log.Printf("[Pricing] Failed to save cache: %v", err)
	}

	// Calculate and save hash
	hash := sha256.Sum256(body)
	hashStr := hex.EncodeToString(hash[:])
	hashFile := filepath.Join(cacheDir, "pricing.sha256")
	if err := os.WriteFile(hashFile, []byte(hashStr), 0644); err != nil {
		log.Printf("[Pricing] Failed to save hash: %v", err)
	}

	// Update in-memory data
	s.mu.Lock()
	s.pricingData = data
	s.lastUpdated = time.Now()
	s.localHash = hashStr
	s.mu.Unlock()

	// Sync to database
	go s.syncToDatabase()

	log.Printf("[Pricing] Downloaded %d models successfully", len(data))
	return nil
}

// loadFromFile loads pricing from a local file
func (s *PricingService) loadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	pricingData, err := s.parsePricing(data)
	if err != nil {
		return fmt.Errorf("parse pricing: %w", err)
	}

	// Calculate hash
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	s.mu.Lock()
	s.pricingData = pricingData
	s.localHash = hashStr
	if info, err := os.Stat(filePath); err == nil {
		s.lastUpdated = info.ModTime()
	}
	s.mu.Unlock()

	return nil
}

// parsePricing parses LiteLLM pricing JSON
func (s *PricingService) parsePricing(data []byte) (map[string]*LiteLLMPricing, error) {
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}

	result := make(map[string]*LiteLLMPricing)

	for modelName, rawEntry := range rawData {
		// Skip documentation entries
		if modelName == "sample_spec" {
			continue
		}

		var pricing LiteLLMPricing
		if err := json.Unmarshal(rawEntry, &pricing); err != nil {
			continue
		}

		// Only keep entries with valid pricing
		if pricing.InputCostPerToken > 0 || pricing.OutputCostPerToken > 0 {
			result[strings.ToLower(modelName)] = &pricing
		}
	}

	return result, nil
}

// syncToDatabase syncs pricing to database
func (s *PricingService) syncToDatabase() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	log.Println("[Pricing] Syncing to database...")

	synced := 0
	for modelName, pricing := range s.pricingData {
		// Only sync Codex models
		if !strings.Contains(modelName, "codex") && !strings.Contains(modelName, "gpt-5") {
			continue
		}

		// Check if model exists
		var existing models.ModelPricing
		result := database.DB.Where("model_name = ?", modelName).First(&existing)

		if result.Error != nil {
			// Create new pricing entry
			newPricing := models.ModelPricing{
				ModelName:        modelName,
				InputPricePer1k:  pricing.InputCostPerToken * 1000000, // Convert to per 1K
				OutputPricePer1k: pricing.OutputCostPerToken * 1000000,
				MarkupMultiplier: 1.5,
			}
			if err := database.DB.Create(&newPricing).Error; err != nil {
				log.Printf("[Pricing] Failed to create %s: %v", modelName, err)
				continue
			}
			synced++
		}
	}

	log.Printf("[Pricing] Synced %d models to database", synced)
}

// startUpdater starts the background update task
func (s *PricingService) startUpdater() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.checkAndUpdate(); err != nil {
					log.Printf("[Pricing] Update check failed: %v", err)
				}
			case <-s.stopCh:
				return
			}
		}
	}()

	log.Printf("[Pricing] Background updater started (check every %v)", checkInterval)
}

// checkAndUpdate checks if update is needed and downloads if necessary
func (s *PricingService) checkAndUpdate() error {
	s.mu.RLock()
	lastUpdate := s.lastUpdated
	s.mu.RUnlock()

	// Check if it's time to update
	if time.Since(lastUpdate) < updateInterval {
		return nil
	}

	log.Println("[Pricing] Update interval reached, downloading...")
	return s.downloadPricing()
}

// GetModelPricing returns pricing for a model (with fuzzy matching)
func (s *PricingService) GetModelPricing(modelName string) *LiteLLMPricing {
	s.mu.RLock()
	defer s.mu.RUnlock()

	modelLower := strings.ToLower(strings.TrimSpace(modelName))

	// Direct match
	if pricing, ok := s.pricingData[modelLower]; ok {
		return pricing
	}

	// Fuzzy match for Codex models
	if strings.Contains(modelLower, "codex") {
		// Try base model
		if strings.Contains(modelLower, "5.1") {
			if pricing, ok := s.pricingData["gpt-5.1-codex"]; ok {
				return pricing
			}
		}
		if strings.Contains(modelLower, "5.2") {
			if pricing, ok := s.pricingData["gpt-5.2-codex"]; ok {
				return pricing
			}
		}
	}

	return nil
}

// GetStatus returns service status
func (s *PricingService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"model_count":  len(s.pricingData),
		"last_updated": s.lastUpdated,
		"local_hash":   s.localHash[:min(8, len(s.localHash))],
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
