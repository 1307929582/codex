package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"codex-gateway/internal/config"
	"codex-gateway/internal/database"
	"codex-gateway/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var httpClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

type OpenAIRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]interface{} `json:"messages"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
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
	} `json:"choices"`
}

func ProxyHandler(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	apiKey := c.MustGet("api_key").(models.APIKey)

	var req OpenAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Stream = false

	startTime := time.Now()

	upstreamResp, err := forwardToOpenAI(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("upstream error: %v", err)})
		return
	}

	latencyMs := int(time.Since(startTime).Milliseconds())

	cost, err := calculateCost(req.Model, upstreamResp.Usage.PromptTokens, upstreamResp.Usage.CompletionTokens)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("pricing error: %v", err)})
		return
	}

	if err := recordUsageAndBill(user.ID, apiKey.ID, req.Model, upstreamResp.Usage.PromptTokens, upstreamResp.Usage.CompletionTokens, cost, latencyMs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "billing transaction failed"})
		return
	}

	c.JSON(http.StatusOK, upstreamResp)
}

func forwardToOpenAI(ctx context.Context, req OpenAIRequest) (*OpenAIResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", config.AppConfig.OpenAIBaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.AppConfig.OpenAIAPIKey)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
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
