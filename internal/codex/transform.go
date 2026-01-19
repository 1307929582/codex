package codex

import (
	"strings"
)

// Codex CLI Instructions embedded
const codexCLIInstructions = `You are Claude Code, Anthropic's official CLI for Claude.

You are an interactive CLI tool that helps users with software engineering tasks.`

// Model name mapping for Codex variants
var codexModelMap = map[string]string{
	"gpt-5.1-codex":             "gpt-5.1-codex",
	"gpt-5.1-codex-low":         "gpt-5.1-codex",
	"gpt-5.1-codex-medium":      "gpt-5.1-codex",
	"gpt-5.1-codex-high":        "gpt-5.1-codex",
	"gpt-5.1-codex-max":         "gpt-5.1-codex-max",
	"gpt-5.1-codex-max-low":     "gpt-5.1-codex-max",
	"gpt-5.1-codex-max-medium":  "gpt-5.1-codex-max",
	"gpt-5.1-codex-max-high":    "gpt-5.1-codex-max",
	"gpt-5.2":                   "gpt-5.2",
	"gpt-5.2-codex":             "gpt-5.2-codex",
	"gpt-5.2-codex-low":         "gpt-5.2-codex",
	"gpt-5.2-codex-medium":      "gpt-5.2-codex",
	"gpt-5.2-codex-high":        "gpt-5.2-codex",
	"gpt-5.1-codex-mini":        "gpt-5.1-codex-mini",
	"gpt-5.1":                   "gpt-5.1",
	"gpt-5-codex":               "gpt-5.1-codex",
	"codex-mini-latest":         "gpt-5.1-codex-mini",
	"gpt-5":                     "gpt-5.1",
}

// TransformRequest applies Codex-specific transformations to the request
func TransformRequest(reqBody map[string]interface{}) bool {
	modified := false

	// Normalize model name
	if model, ok := reqBody["model"].(string); ok {
		normalized := NormalizeModelName(model)
		if normalized != model {
			reqBody["model"] = normalized
			modified = true
		}
	}

	// Force store=false for OAuth compatibility
	if store, ok := reqBody["store"].(bool); !ok || store {
		reqBody["store"] = false
		modified = true
	}

	// Inject Codex CLI instructions if not present
	if instructions, ok := reqBody["instructions"].(string); !ok || strings.TrimSpace(instructions) == "" {
		reqBody["instructions"] = codexCLIInstructions
		modified = true
	}

	// Remove unsupported parameters
	if _, ok := reqBody["max_output_tokens"]; ok {
		delete(reqBody, "max_output_tokens")
		modified = true
	}
	if _, ok := reqBody["max_completion_tokens"]; ok {
		delete(reqBody, "max_completion_tokens")
		modified = true
	}

	return modified
}

// NormalizeModelName normalizes Codex model names
func NormalizeModelName(model string) string {
	if model == "" {
		return "gpt-5.1"
	}

	// Extract model ID from path format
	modelID := model
	if strings.Contains(modelID, "/") {
		parts := strings.Split(modelID, "/")
		modelID = parts[len(parts)-1]
	}

	// Check direct mapping
	if mapped, ok := codexModelMap[modelID]; ok {
		return mapped
	}

	// Fuzzy matching
	normalized := strings.ToLower(modelID)

	if strings.Contains(normalized, "gpt-5.2-codex") || strings.Contains(normalized, "gpt 5.2 codex") {
		return "gpt-5.2-codex"
	}
	if strings.Contains(normalized, "gpt-5.2") || strings.Contains(normalized, "gpt 5.2") {
		return "gpt-5.2"
	}
	if strings.Contains(normalized, "gpt-5.1-codex-max") || strings.Contains(normalized, "gpt 5.1 codex max") {
		return "gpt-5.1-codex-max"
	}
	if strings.Contains(normalized, "gpt-5.1-codex-mini") || strings.Contains(normalized, "gpt 5.1 codex mini") {
		return "gpt-5.1-codex-mini"
	}
	if strings.Contains(normalized, "gpt-5.1-codex") || strings.Contains(normalized, "gpt 5.1 codex") {
		return "gpt-5.1-codex"
	}
	if strings.Contains(normalized, "gpt-5.1") || strings.Contains(normalized, "gpt 5.1") {
		return "gpt-5.1"
	}
	if strings.Contains(normalized, "codex") {
		return "gpt-5.1-codex"
	}
	if strings.Contains(normalized, "gpt-5") || strings.Contains(normalized, "gpt 5") {
		return "gpt-5.1"
	}

	return "gpt-5.1"
}
