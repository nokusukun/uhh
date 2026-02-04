package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// ModelInfo represents information about an available model
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	OwnedBy     string `json:"owned_by,omitempty"`
}

// ModelLister is an interface for providers that can list available models
type ModelLister interface {
	ListModels(ctx context.Context, apiKey string) ([]ModelInfo, error)
}

// OpenAI Models API response
type openAIModelsResponse struct {
	Data []struct {
		ID      string `json:"id"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// ListModelsOpenAI fetches available models from OpenAI API
func ListModelsOpenAI(ctx context.Context, apiKey, baseURL string) ([]ModelInfo, error) {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result openAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]ModelInfo, 0, len(result.Data))
	for _, m := range result.Data {
		// Filter to only include chat/completion models
		if isRelevantModel(m.ID) {
			models = append(models, ModelInfo{
				ID:      m.ID,
				Name:    m.ID,
				OwnedBy: m.OwnedBy,
			})
		}
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].ID < models[j].ID
	})

	return models, nil
}

// ListModelsGemini fetches available models from Google Gemini API
func ListModelsGemini(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name                       string   `json:"name"`
			DisplayName                string   `json:"displayName"`
			Description                string   `json:"description"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]ModelInfo, 0)
	for _, m := range result.Models {
		// Only include models that support generateContent
		supportsContent := false
		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				supportsContent = true
				break
			}
		}
		if !supportsContent {
			continue
		}

		// Extract model ID from name (e.g., "models/gemini-pro" -> "gemini-pro")
		id := m.Name
		if strings.HasPrefix(id, "models/") {
			id = strings.TrimPrefix(id, "models/")
		}

		models = append(models, ModelInfo{
			ID:          id,
			Name:        m.DisplayName,
			Description: m.Description,
		})
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].ID < models[j].ID
	})

	return models, nil
}

// ListModelsDeepSeek fetches available models from DeepSeek API
func ListModelsDeepSeek(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	return ListModelsOpenAI(ctx, apiKey, deepseekBaseURL)
}

// ListModelsKimi fetches available models from Kimi/Moonshot API
func ListModelsKimi(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	// Determine which API to use based on key prefix
	baseURL := kimiBaseURL
	var httpClient *http.Client

	if strings.HasPrefix(apiKey, "sk-kimi-") {
		baseURL = kimiCodingBaseURL
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &codingAgentTransport{
				base: http.DefaultTransport,
			},
		}
	} else {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result openAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]ModelInfo, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, ModelInfo{
			ID:      m.ID,
			Name:    m.ID,
			OwnedBy: m.OwnedBy,
		})
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].ID < models[j].ID
	})

	return models, nil
}

// ListModelsGLM fetches available models from GLM/Zhipu API
func ListModelsGLM(ctx context.Context, apiKey string) ([]ModelInfo, error) {
	return ListModelsOpenAI(ctx, apiKey, glmBaseURL)
}

// ListModels fetches available models for a given provider
func ListModels(ctx context.Context, providerName, apiKey, baseURL string) ([]ModelInfo, error) {
	switch providerName {
	case "openai":
		return ListModelsOpenAI(ctx, apiKey, baseURL)
	case "gemini":
		return ListModelsGemini(ctx, apiKey)
	case "deepseek":
		return ListModelsDeepSeek(ctx, apiKey)
	case "kimi":
		return ListModelsKimi(ctx, apiKey)
	case "glm":
		return ListModelsGLM(ctx, apiKey)
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}
}

// isRelevantModel filters models to only include chat/completion models
func isRelevantModel(id string) bool {
	// Include GPT models
	if strings.Contains(id, "gpt") {
		return true
	}
	// Include chat models
	if strings.Contains(id, "chat") {
		return true
	}
	// Include DeepSeek models
	if strings.Contains(id, "deepseek") {
		return true
	}
	// Include Claude models (for compatible endpoints)
	if strings.Contains(id, "claude") {
		return true
	}
	// Include o1/o3 reasoning models
	if strings.HasPrefix(id, "o1") || strings.HasPrefix(id, "o3") {
		return true
	}
	// Exclude embedding, moderation, whisper, tts, dall-e models
	if strings.Contains(id, "embedding") ||
		strings.Contains(id, "moderation") ||
		strings.Contains(id, "whisper") ||
		strings.Contains(id, "tts") ||
		strings.Contains(id, "dall-e") ||
		strings.Contains(id, "davinci") ||
		strings.Contains(id, "babbage") ||
		strings.Contains(id, "curie") ||
		strings.Contains(id, "ada") {
		return false
	}
	return false
}

// GetDefaultModelsForProvider returns a list of commonly used models for a provider
// This is used as a fallback when API fetching fails
func GetDefaultModelsForProvider(providerName string) []ModelInfo {
	switch providerName {
	case "openai":
		return []ModelInfo{
			{ID: "gpt-4o", Name: "GPT-4o", Description: "Most capable GPT-4 model"},
			{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Description: "Smaller, faster GPT-4o"},
			{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", Description: "GPT-4 Turbo with vision"},
			{ID: "gpt-4", Name: "GPT-4", Description: "Original GPT-4"},
			{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Description: "Fast and efficient"},
			{ID: "o1-preview", Name: "o1 Preview", Description: "Reasoning model (preview)"},
			{ID: "o1-mini", Name: "o1 Mini", Description: "Smaller reasoning model"},
		}
	case "gemini":
		return []ModelInfo{
			{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", Description: "Latest fast model"},
			{ID: "gemini-2.0-flash-thinking-exp", Name: "Gemini 2.0 Flash Thinking", Description: "With reasoning"},
			{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", Description: "Most capable"},
			{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", Description: "Fast and efficient"},
			{ID: "gemini-pro", Name: "Gemini Pro", Description: "Original Gemini Pro"},
		}
	case "deepseek":
		return []ModelInfo{
			{ID: "deepseek-chat", Name: "DeepSeek Chat", Description: "General chat model"},
			{ID: "deepseek-coder", Name: "DeepSeek Coder", Description: "Code generation"},
			{ID: "deepseek-reasoner", Name: "DeepSeek Reasoner", Description: "Reasoning model"},
		}
	case "kimi":
		return []ModelInfo{
			{ID: "moonshot-v1-8k", Name: "Moonshot V1 8K", Description: "8K context"},
			{ID: "moonshot-v1-32k", Name: "Moonshot V1 32K", Description: "32K context"},
			{ID: "moonshot-v1-128k", Name: "Moonshot V1 128K", Description: "128K context"},
			{ID: "kimi-coding/k2p5", Name: "Kimi Coding K2P5", Description: "KimiCoding model"},
		}
	case "glm":
		return []ModelInfo{
			{ID: "glm-4", Name: "GLM-4", Description: "Most capable GLM model"},
			{ID: "glm-4-flash", Name: "GLM-4 Flash", Description: "Fast GLM-4"},
			{ID: "glm-3-turbo", Name: "GLM-3 Turbo", Description: "Efficient model"},
		}
	default:
		return nil
	}
}
