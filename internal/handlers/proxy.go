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
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
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

	// Apply Codex transformations
	codex.TransformRequest(reqBody)

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
		handleStreamingRequest(c, user, apiKey, reqBody, model, startTime)
	} else {
		handleNonStreamingRequest(c, user, apiKey, reqBody, model, startTime)
	}
}

func handleStreamingRequest(c *gin.Context, user models.User, apiKey models.APIKey, reqBody map[string]interface{}, model string, startTime time.Time) {
	// Get upstream config
	var settings models.SystemSettings
	if err := database.DB.First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}

	if settings.OpenAIAPIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upstream API key not configured"})
		return
	}

	baseURL := settings.OpenAIBaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	// Ensure stream=true for upstream and request usage info
	reqBody["stream"] = true
	reqBody["stream_options"] = map[string]interface{}{"include_usage": true}

	// Marshal request
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request"})
		return
	}

	// Create upstream request
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", baseURL+"/chat/completions", bytes.NewBuffer(reqBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+settings.OpenAIAPIKey)
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
				if chunk.Usage.TotalTokens > 0 {
					lastUsage.PromptTokens = chunk.Usage.PromptTokens
					lastUsage.CompletionTokens = chunk.Usage.CompletionTokens
					lastUsage.TotalTokens = chunk.Usage.TotalTokens
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
		cost, err := calculateCost(model, lastUsage.PromptTokens, lastUsage.CompletionTokens)
		if err == nil {
			_ = recordUsageAndBill(user.ID, apiKey.ID, model, lastUsage.PromptTokens, lastUsage.CompletionTokens, cost, latencyMs)
		}
	} else if streamedChunks > 0 {
		// Fallback: estimate tokens if usage info not available
		// Rough estimate: ~4 chars per token, ~100 chars per chunk
		estimatedTokens := streamedChunks * 25
		estimatedInput := estimatedTokens / 10  // Assume 10% input
		estimatedOutput := estimatedTokens - estimatedInput

		cost, err := calculateCost(model, estimatedInput, estimatedOutput)
		if err == nil {
			_ = recordUsageAndBill(user.ID, apiKey.ID, model, estimatedInput, estimatedOutput, cost, latencyMs)
		}
	}
}

func handleNonStreamingRequest(c *gin.Context, user models.User, apiKey models.APIKey, reqBody map[string]interface{}, model string, startTime time.Time) {
	// Force stream=false for non-streaming
	reqBody["stream"] = false

	upstreamResp, err := forwardToUpstream(c.Request.Context(), reqBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("upstream error: %v", err)})
		return
	}

	latencyMs := int(time.Since(startTime).Milliseconds())

	cost, err := calculateCost(model, upstreamResp.Usage.PromptTokens, upstreamResp.Usage.CompletionTokens)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("pricing error: %v", err)})
		return
	}

	if err := recordUsageAndBill(user.ID, apiKey.ID, model, upstreamResp.Usage.PromptTokens, upstreamResp.Usage.CompletionTokens, cost, latencyMs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "billing failed"})
		return
	}

	c.JSON(http.StatusOK, upstreamResp)
}

func forwardToUpstream(ctx context.Context, reqBody map[string]interface{}) (*OpenAIResponse, error) {
	var settings models.SystemSettings
	if err := database.DB.First(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	if settings.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("upstream API key not configured")
	}

	baseURL := settings.OpenAIBaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+settings.OpenAIAPIKey)

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
	var pricing models.ModelPricing
	if err := database.DB.Where("model_name = ?", model).First(&pricing).Error; err != nil {
		return 0, fmt.Errorf("pricing not found for model: %s", model)
	}

	inputCost := (float64(inputTokens) / 1000.0) * pricing.InputPricePer1k
	outputCost := (float64(outputTokens) / 1000.0) * pricing.OutputPricePer1k
	return (inputCost + outputCost) * pricing.MarkupMultiplier, nil
}

func recordUsageAndBill(userID uuid.UUID, apiKeyID uint, model string, inputTokens, outputTokens int, cost float64, latencyMs int) error {
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
