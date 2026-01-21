package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"codex-gateway/internal/billing"
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
		PromptTokens       int `json:"prompt_tokens"`
		CompletionTokens   int `json:"completion_tokens"`
		TotalTokens        int `json:"total_tokens"`
		PromptTokenDetails struct {
			CachedTokens        int `json:"cached_tokens"`
			CacheReadTokens     int `json:"cache_read_tokens"`
			CacheCreationTokens int `json:"cache_creation_tokens"`
		} `json:"prompt_tokens_details"`

		// Codex/Responses API fields
		InputTokens       int `json:"input_tokens"`
		OutputTokens      int `json:"output_tokens"`
		InputTokenDetails struct {
			CachedTokens        int `json:"cached_tokens"`
			CacheReadTokens     int `json:"cache_read_tokens"`
			CacheCreationTokens int `json:"cache_creation_tokens"`
		} `json:"input_tokens_details"`
		InputTokenDetailsAlt struct {
			CachedTokens        int `json:"cached_tokens"`
			CacheReadTokens     int `json:"cache_read_tokens"`
			CacheCreationTokens int `json:"cache_creation_tokens"`
		} `json:"input_token_details"`
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

	// Pre-flight check: user must have balance OR active package
	today := database.GetToday()
	var activePackage models.UserPackage
	hasActivePackage := database.DB.Where("user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
		user.ID, "active", today, today).
		First(&activePackage).Error == nil

	if user.Balance <= 0 && !hasActivePackage {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance or active package"})
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

func selectUpstreamForUser(userID uuid.UUID) (*models.CodexUpstream, error) {
	upstreamObj, err := upstream.GetSelector().SelectForUser(userID)
	if err == nil {
		return upstreamObj, nil
	}

	var settings models.SystemSettings
	if err := database.DB.First(&settings).Error; err != nil {
		return nil, err
	}
	if settings.OpenAIAPIKey == "" {
		return nil, err
	}
	baseURL := settings.OpenAIBaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &models.CodexUpstream{
		Name:    "Settings Fallback Upstream",
		BaseURL: baseURL,
		APIKey:  settings.OpenAIAPIKey,
		Status:  "active",
	}, nil
}

func handleStreamingRequest(c *gin.Context, user models.User, apiKey models.APIKey, reqBody map[string]interface{}, model string, requestPath string, startTime time.Time) {
	// Select upstream for this user (consistent hashing for session affinity)
	upstreamObj, err := selectUpstreamForUser(user.ID)
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
		PromptTokens        int
		CompletionTokens    int
		CachedTokens        int
		CacheCreationTokens int
		TotalTokens         int
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// Track streaming state
	streamedChunks := 0
	outputBytes := 0
	clientDisconnected := false

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
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

			// Try to parse as Codex response.completed event first
			var codexEvent struct {
				Type     string `json:"type"`
				Response struct {
					Usage struct {
						InputTokens       int `json:"input_tokens"`
						OutputTokens      int `json:"output_tokens"`
						InputTokenDetails struct {
							CachedTokens        int `json:"cached_tokens"`
							CacheReadTokens     int `json:"cache_read_tokens"`
							CacheCreationTokens int `json:"cache_creation_tokens"`
						} `json:"input_tokens_details"`
						InputTokenDetailsAlt struct {
							CachedTokens        int `json:"cached_tokens"`
							CacheReadTokens     int `json:"cache_read_tokens"`
							CacheCreationTokens int `json:"cache_creation_tokens"`
						} `json:"input_token_details"`
					} `json:"usage"`
				} `json:"response"`
			}

			if err := json.Unmarshal([]byte(data), &codexEvent); err == nil && codexEvent.Type == "response.completed" {
				// Codex/Responses API format
				cacheReadTokens := codexEvent.Response.Usage.InputTokenDetails.CacheReadTokens
				if cacheReadTokens == 0 {
					cacheReadTokens = codexEvent.Response.Usage.InputTokenDetails.CachedTokens
				}
				if cacheReadTokens == 0 {
					cacheReadTokens = codexEvent.Response.Usage.InputTokenDetailsAlt.CacheReadTokens
					if cacheReadTokens == 0 {
						cacheReadTokens = codexEvent.Response.Usage.InputTokenDetailsAlt.CachedTokens
					}
				}
				cacheCreationTokens := codexEvent.Response.Usage.InputTokenDetails.CacheCreationTokens
				if cacheCreationTokens == 0 {
					cacheCreationTokens = codexEvent.Response.Usage.InputTokenDetailsAlt.CacheCreationTokens
				}

				lastUsage.PromptTokens = codexEvent.Response.Usage.InputTokens
				lastUsage.CompletionTokens = codexEvent.Response.Usage.OutputTokens
				lastUsage.CachedTokens = cacheReadTokens
				lastUsage.CacheCreationTokens = cacheCreationTokens
				lastUsage.TotalTokens = resolveTotalTokens(lastUsage.PromptTokens, lastUsage.CompletionTokens, lastUsage.CachedTokens, lastUsage.CacheCreationTokens)
				continue
			}

			// Track streamed output length for fallback billing
			var deltaEvent struct {
				Type  string          `json:"type"`
				Delta json.RawMessage `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &deltaEvent); err == nil {
				switch deltaEvent.Type {
				case "response.output_text.delta":
					var deltaText string
					if err := json.Unmarshal(deltaEvent.Delta, &deltaText); err == nil {
						outputBytes += len(deltaText)
					}
				case "response.content_part.delta":
					var delta struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}
					if err := json.Unmarshal(deltaEvent.Delta, &delta); err == nil {
						outputBytes += len(delta.Text)
					}
				}
			}

			// Try ChatGPT format
			var chunk OpenAIResponse
			if err := json.Unmarshal([]byte(data), &chunk); err == nil {
				if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
					outputBytes += len(chunk.Choices[0].Delta.Content)
				}
				if chunk.Usage.TotalTokens > 0 {
					lastUsage.PromptTokens = chunk.Usage.PromptTokens
					lastUsage.CompletionTokens = chunk.Usage.CompletionTokens
					cacheReadTokens := chunk.Usage.PromptTokenDetails.CacheReadTokens
					if cacheReadTokens == 0 {
						cacheReadTokens = chunk.Usage.PromptTokenDetails.CachedTokens
					}
					cacheCreationTokens := chunk.Usage.PromptTokenDetails.CacheCreationTokens
					lastUsage.CachedTokens = cacheReadTokens
					lastUsage.CacheCreationTokens = cacheCreationTokens
					lastUsage.TotalTokens = resolveTotalTokens(lastUsage.PromptTokens, lastUsage.CompletionTokens, lastUsage.CachedTokens, lastUsage.CacheCreationTokens)
				} else if chunk.Usage.InputTokens > 0 || chunk.Usage.OutputTokens > 0 {
					// Direct usage format (non-event)
					lastUsage.PromptTokens = chunk.Usage.InputTokens
					lastUsage.CompletionTokens = chunk.Usage.OutputTokens
					cacheReadTokens := chunk.Usage.InputTokenDetails.CacheReadTokens
					if cacheReadTokens == 0 {
						cacheReadTokens = chunk.Usage.InputTokenDetails.CachedTokens
					}
					if cacheReadTokens == 0 {
						cacheReadTokens = chunk.Usage.InputTokenDetailsAlt.CacheReadTokens
						if cacheReadTokens == 0 {
							cacheReadTokens = chunk.Usage.InputTokenDetailsAlt.CachedTokens
						}
					}
					cacheCreationTokens := chunk.Usage.InputTokenDetails.CacheCreationTokens
					if cacheCreationTokens == 0 {
						cacheCreationTokens = chunk.Usage.InputTokenDetailsAlt.CacheCreationTokens
					}
					lastUsage.CachedTokens = cacheReadTokens
					lastUsage.CacheCreationTokens = cacheCreationTokens
					lastUsage.TotalTokens = resolveTotalTokens(lastUsage.PromptTokens, lastUsage.CompletionTokens, lastUsage.CachedTokens, lastUsage.CacheCreationTokens)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		// Stream already started, can't send JSON error
		// But we should still try to bill for what was sent
		log.Printf("[Proxy] Stream scan error: %v", err)
	}

	// Calculate latency
	latencyMs := int(time.Since(startTime).Milliseconds())

	// Bill user after stream completes
	if lastUsage.TotalTokens > 0 {
		cost, err := calculateCostWithCache(model, lastUsage.PromptTokens, lastUsage.CompletionTokens, lastUsage.CachedTokens, lastUsage.CacheCreationTokens)
		if err == nil {
			_ = recordUsageAndBill(user.ID, apiKey.ID, model, lastUsage.PromptTokens, lastUsage.CompletionTokens, lastUsage.CachedTokens, lastUsage.CacheCreationTokens, cost, latencyMs)
		}
	} else if outputBytes > 0 || streamedChunks > 0 {
		// Fallback: estimate tokens if usage info not available
		// Approximate: ~4 bytes per token for output text, assume 10% input
		estimatedOutput := outputBytes / 4
		if outputBytes%4 != 0 {
			estimatedOutput++
		}
		if estimatedOutput == 0 && streamedChunks > 0 {
			estimatedOutput = streamedChunks * 10
		}
		estimatedInput := estimatedOutput / 10

		cost, err := calculateCostWithCache(model, estimatedInput, estimatedOutput, 0, 0)
		if err == nil {
			_ = recordUsageAndBill(user.ID, apiKey.ID, model, estimatedInput, estimatedOutput, 0, 0, cost, latencyMs)
		}
	}
}

func handleNonStreamingRequest(c *gin.Context, user models.User, apiKey models.APIKey, reqBody map[string]interface{}, model string, requestPath string, startTime time.Time) {
	// Select upstream for this user (consistent hashing for session affinity)
	upstreamObj, err := selectUpstreamForUser(user.ID)
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
	cachedTokens := upstreamResp.Usage.PromptTokenDetails.CacheReadTokens
	if cachedTokens == 0 {
		cachedTokens = upstreamResp.Usage.PromptTokenDetails.CachedTokens
	}
	cacheCreationTokens := upstreamResp.Usage.PromptTokenDetails.CacheCreationTokens

	// If ChatGPT fields are empty, try Codex fields
	if inputTokens == 0 && outputTokens == 0 {
		inputTokens = upstreamResp.Usage.InputTokens
		outputTokens = upstreamResp.Usage.OutputTokens
		cachedTokens = upstreamResp.Usage.InputTokenDetails.CacheReadTokens
		if cachedTokens == 0 {
			cachedTokens = upstreamResp.Usage.InputTokenDetails.CachedTokens
		}
		if cachedTokens == 0 {
			cachedTokens = upstreamResp.Usage.InputTokenDetailsAlt.CacheReadTokens
			if cachedTokens == 0 {
				cachedTokens = upstreamResp.Usage.InputTokenDetailsAlt.CachedTokens
			}
		}
		cacheCreationTokens = upstreamResp.Usage.InputTokenDetails.CacheCreationTokens
		if cacheCreationTokens == 0 {
			cacheCreationTokens = upstreamResp.Usage.InputTokenDetailsAlt.CacheCreationTokens
		}
	} else if cachedTokens == 0 {
		cachedTokens = upstreamResp.Usage.InputTokenDetails.CacheReadTokens
		if cachedTokens == 0 {
			cachedTokens = upstreamResp.Usage.InputTokenDetails.CachedTokens
		}
		if cachedTokens == 0 {
			cachedTokens = upstreamResp.Usage.InputTokenDetailsAlt.CacheReadTokens
			if cachedTokens == 0 {
				cachedTokens = upstreamResp.Usage.InputTokenDetailsAlt.CachedTokens
			}
		}
	}

	if cacheCreationTokens == 0 {
		cacheCreationTokens = upstreamResp.Usage.InputTokenDetails.CacheCreationTokens
		if cacheCreationTokens == 0 {
			cacheCreationTokens = upstreamResp.Usage.InputTokenDetailsAlt.CacheCreationTokens
		}
	}

	cost, err := calculateCostWithCache(model, inputTokens, outputTokens, cachedTokens, cacheCreationTokens)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("pricing error: %v", err)})
		return
	}

	if err := recordUsageAndBill(user.ID, apiKey.ID, model, inputTokens, outputTokens, cachedTokens, cacheCreationTokens, cost, latencyMs); err != nil {
		if strings.Contains(err.Error(), "insufficient balance") ||
			strings.Contains(err.Error(), "api key quota exceeded") ||
			strings.Contains(err.Error(), "daily usage limit exceeded") {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
			return
		}
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

func resolveBillableInputTokens(inputTokens, cacheReadTokens, cacheCreationTokens int) int {
	if inputTokens <= 0 {
		return 0
	}
	billable := inputTokens - cacheReadTokens - cacheCreationTokens
	if billable < 0 {
		return 0
	}
	return billable
}

func resolveTotalTokens(inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens int) int {
	total := inputTokens + outputTokens
	cacheInput := cacheReadTokens + cacheCreationTokens
	if cacheInput > inputTokens {
		total = cacheInput + outputTokens
	}
	return total
}

func calculateCost(model string, inputTokens, outputTokens int) (float64, error) {
	return calculateCostWithCache(model, inputTokens, outputTokens, 0, 0)
}

func calculateCostWithCache(model string, inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens int) (float64, error) {
	var pricing models.ModelPricing
	if err := database.DB.Where("model_name = ?", model).First(&pricing).Error; err != nil {
		return 0, fmt.Errorf("pricing not found for model: %s", model)
	}

	// Calculate costs for each token type
	// Note: cached_tokens in Codex API = cache_read_tokens (tokens read from cache)
	// Cache read/creation tokens are billed at discounted rates.
	billableInputTokens := resolveBillableInputTokens(inputTokens, cacheReadTokens, cacheCreationTokens)
	inputCost := (float64(billableInputTokens) / 1000.0) * pricing.InputPricePer1k
	cacheReadCost := (float64(cacheReadTokens) / 1000.0) * pricing.CacheReadPricePer1k
	cacheCreateCost := (float64(cacheCreationTokens) / 1000.0) * pricing.CacheCreationPricePer1k
	outputCost := (float64(outputTokens) / 1000.0) * pricing.OutputPricePer1k

	return (inputCost + cacheReadCost + cacheCreateCost + outputCost) * pricing.MarkupMultiplier, nil
}

func recordUsageAndBill(userID uuid.UUID, apiKeyID uint, model string, inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens int, cost float64, latencyMs int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Use new billing logic that supports package quota
		if err := billing.DeductCost(tx, userID, cost); err != nil {
			return err
		}

		totalTokens := resolveTotalTokens(inputTokens, outputTokens, cacheReadTokens, cacheCreationTokens)

		log := models.UsageLog{
			UserID:              userID,
			APIKeyID:            apiKeyID,
			Model:               model,
			InputTokens:         inputTokens,
			OutputTokens:        outputTokens,
			CachedTokens:        cacheReadTokens,
			CacheCreationTokens: cacheCreationTokens,
			TotalTokens:         totalTokens,
			Cost:                cost,
			LatencyMs:           latencyMs,
			StatusCode:          http.StatusOK,
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		result := tx.Model(&models.APIKey{}).
			Where("id = ? AND (quota_limit IS NULL OR total_usage + ? <= quota_limit)", apiKeyID, totalTokens).
			UpdateColumn("total_usage", gorm.Expr("total_usage + ?", totalTokens))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("api key quota exceeded")
		}

		return nil
	})
}
