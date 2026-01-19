package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"codex-gateway/internal/codex"
	"codex-gateway/internal/database"
	"codex-gateway/internal/models"
	"codex-gateway/internal/upstream"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var httpClient = &http.Client{
	Timeout: 120 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

type OpenAIRequest struct {
	Model        string                   `json:"model"`
	Messages     []map[string]interface{} `json:"messages"`
	Temperature  float64                  `json:"temperature,omitempty"`
	MaxTokens    int                      `json:"max_tokens,omitempty"`
	Stream       bool                     `json:"stream,omitempty"`
	Instructions string                   `json:"instructions,omitempty"`
	Store        bool                     `json:"store,omitempty"`
}

type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		// ChatGPT API fields
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`

		// Codex/Responses API fields
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		CachedTokens int `json:"cached_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
		Delta        struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta,omitempty"`
	} `json:"choices"`
}

func ProxyHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	apiKey := c.MustGet("api_key").(models.APIKey)

	// Get the original request path (e.g., /chat/completions, /responses, /completions)
	requestPath := c.Request.URL.Path

	// Parse request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	var reqBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// Only apply ChatGPT transformations for /chat/completions endpoint
	// For Codex endpoints (/responses, /completions), pass through as-is
	if strings.HasSuffix(requestPath, "/chat/completions") {
		codex.TransformRequest(reqBody)
	}

	// Pre-flight balance check
	if user.Balance <= 0 {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
		return
	}

	// Get model name for billing
	model, _ := reqBody["model"].(string)
	if model == "" {
		model = "gpt-5.1-codex"
	}

	// Check if streaming is requested
	stream, _ := reqBody["stream"].(bool)

	startTime := time.Now()

	if stream {
		handleStreamingRequest(c, user, apiKey, reqBody, model, requestPath, startTime)
	} else {
		handleNonStreamingRequest(c, user, apiKey, reqBody, model, requestPath, startTime)
	}
}

func handleStreamingRequest(c *gin.Context, user models.User, apiKey models.APIKey, reqBody map[string]interface{}, model string, requestPath string, startTime time.Time) {
	// Select upstream for this user (consistent hashing for session affinity)
	upstreamObj, err := upstream.GetSelector().SelectForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available upstream"})
		return
	}

	baseURL := upstreamObj.BaseURL
	apiKeyStr := upstreamObj.APIKey

	// Ensure stream=true for upstream
	reqBody["stream"] = true

	// Only add stream_options for ChatGPT API (not for Codex/Responses API)
	if strings.HasSuffix(requestPath, "/chat/completions") {
		reqBody["stream_options"] = map[string]interface{}{"include_usage": true}
	}

	// Marshal request
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request"})
		return
	}

	// Create upstream request using the original request path
	upstreamURL := baseURL + requestPath
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", upstreamURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKeyStr)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Send request
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("upstream error: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("upstream returned %d: %s", resp.StatusCode, string(bodyBytes))})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Stream response and collect usage
	var lastUsage struct {
		PromptTokens     int
		CompletionTokens int
		CachedTokens     int
		TotalTokens      int
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// Track streaming state
	streamedChunks := 0
	clientDisconnected := false

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		// Check if client disconnected
		select {
		case <-c.Request.Context().Done():
			clientDisconnected = true
			break
		default:
		}

		if clientDisconnected {
			break
		}

		line := scanner.Text()

		// Forward SSE line
		if _, err := fmt.Fprintf(c.Writer, "%s\n", line); err != nil {
			clientDisconnected = true
			break
		}
		flusher.Flush()
		streamedChunks++

		// Parse for usage information
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				continue
			}

			var chunk OpenAIResponse
			if err := json.Unmarshal([]byte(data), &chunk); err == nil {
				// Try ChatGPT API fields first
				if chunk.Usage.TotalTokens > 0 {
					lastUsage.PromptTokens = chunk.Usage.PromptTokens
					lastUsage.CompletionTokens = chunk.Usage.CompletionTokens
					lastUsage.TotalTokens = chunk.Usage.TotalTokens
				} else if chunk.Usage.InputTokens > 0 || chunk.Usage.OutputTokens > 0 {
					// Codex/Responses API uses different field names
					lastUsage.PromptTokens = chunk.Usage.InputTokens
					lastUsage.CompletionTokens = chunk.Usage.OutputTokens
					lastUsage.CachedTokens = chunk.Usage.CachedTokens
					lastUsage.TotalTokens = chunk.Usage.InputTokens + chunk.Usage.OutputTokens
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		// Stream already started, can't send JSON error
		// But we should still try to bill for what was sent
	}

	// Calculate latency
	latencyMs := int(time.Since(startTime).Milliseconds())

	// Bill user after stream completes
	if lastUsage.TotalTokens > 0 {
		cost, err := calculateCostWithCache(model, lastUsage.PromptTokens, lastUsage.CompletionTokens, lastUsage.CachedTokens)
		if err == nil {
			_ = recordUsageAndBill(user.ID, apiKey.ID, model, lastUsage.PromptTokens, lastUsage.CompletionTokens, lastUsage.CachedTokens, cost, latencyMs)
		}
	} else if streamedChunks > 0 {
		// Fallback: estimate tokens if usage info not available
		// Rough estimate: ~4 chars per token, ~100 chars per chunk
		estimatedTokens := streamedChunks * 25
		estimatedInput := estimatedTokens / 10  // Assume 10% input
		estimatedOutput := estimatedTokens - estimatedInput

		cost, err := calculateCostWithCache(model, estimatedInput, estimatedOutput, 0)
		if err == nil {
			_ = recordUsageAndBill(user.ID, apiKey.ID, model, estimatedInput, estimatedOutput, 0, cost, latencyMs)
		}
	}
}

func handleNonStreamingRequest(c *gin.Context, user models.User, apiKey models.APIKey, reqBody map[string]interface{}, model string, requestPath string, startTime time.Time) {
	// Select upstream for this user (consistent hashing for session affinity)
	upstreamObj, err := upstream.GetSelector().SelectForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available upstream"})
		return
	}

	// Force stream=false for non-streaming
	reqBody["stream"] = false

	upstreamResp, err := forwardToUpstream(c.Request.Context(), reqBody, upstreamObj.BaseURL, upstreamObj.APIKey, requestPath)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("upstream error: %v", err)})
		return
	}

	latencyMs := int(time.Since(startTime).Milliseconds())

	// Extract token counts (support both ChatGPT and Codex API formats)
	inputTokens := upstreamResp.Usage.PromptTokens
	outputTokens := upstreamResp.Usage.CompletionTokens
	cachedTokens := 0

	// If ChatGPT fields are empty, try Codex fields
	if inputTokens == 0 && outputTokens == 0 {
		inputTokens = upstreamResp.Usage.InputTokens
		outputTokens = upstreamResp.Usage.OutputTokens
		cachedTokens = upstreamResp.Usage.CachedTokens
	}

	cost, err := calculateCostWithCache(model, inputTokens, outputTokens, cachedTokens)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("pricing error: %v", err)})
		return
	}

	if err := recordUsageAndBill(user.ID, apiKey.ID, model, inputTokens, outputTokens, cachedTokens, cost, latencyMs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "billing failed"})
		return
	}

	c.JSON(http.StatusOK, upstreamResp)
}

func forwardToUpstream(ctx context.Context, reqBody map[string]interface{}, baseURL string, apiKey string, requestPath string) (*OpenAIResponse, error) {
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// Use the original request path
	upstreamURL := baseURL + requestPath
	httpReq, err := http.NewRequestWithContext(ctx, "POST", upstreamURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("upstream returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	return &openAIResp, nil
}

func calculateCost(model string, inputTokens, outputTokens int) (float64, error) {
	return calculateCostWithCache(model, inputTokens, outputTokens, 0)
}

func calculateCostWithCache(model string, inputTokens, outputTokens, cachedTokens int) (float64, error) {
	var pricing models.ModelPricing
	if err := database.DB.Where("model_name = ?", model).First(&pricing).Error; err != nil {
		return 0, fmt.Errorf("pricing not found for model: %s", model)
	}

	// Calculate costs for each token type
	// Note: input_tokens and cached_tokens are separate in Codex API
	// input_tokens = new tokens that need processing (full price)
	// cached_tokens = tokens from cache (discounted price, usually 50% off)
	inputCost := (float64(inputTokens) / 1000.0) * pricing.InputPricePer1k
	cachedCost := (float64(cachedTokens) / 1000.0) * pricing.CachedInputPricePer1k
	outputCost := (float64(outputTokens) / 1000.0) * pricing.OutputPricePer1k

	return (inputCost + cachedCost + outputCost) * pricing.MarkupMultiplier, nil
}

func recordUsageAndBill(userID uuid.UUID, apiKeyID uint, model string, inputTokens, outputTokens, cachedTokens int, cost float64, latencyMs int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec("UPDATE users SET balance = balance - ? WHERE id = ? AND balance >= ?", cost, userID, cost)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("insufficient balance or user not found")
		}

		log := models.UsageLog{
			UserID:       userID,
			APIKeyID:     apiKeyID,
			Model:        model,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			CachedTokens: cachedTokens,
			TotalTokens:  inputTokens + outputTokens,
			Cost:         cost,
			LatencyMs:    latencyMs,
			StatusCode:   http.StatusOK,
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.APIKey{}).Where("id = ?", apiKeyID).UpdateColumn("total_usage", gorm.Expr("total_usage + ?", inputTokens+outputTokens)).Error; err != nil {
			return err
		}

		return nil
	})
}
