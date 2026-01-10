package gemini

import (
	"context"

	"google.golang.org/genai"
)

type Client struct {
	client    *genai.Client
	ModelName string
}

type ChatMessage struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

func NewClient(ctx context.Context, apiKey string, modelName string) (*Client, error) {
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		client:    c,
		ModelName: modelName,
	}, nil
}

func (c *Client) Close() {
	// The new SDK doesn't require explicit closing
}

// ListModels returns a list of available text and multimodal generation models
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	var models []string

	// Use All() iterator for automatic pagination
	for m, err := range c.client.Models.All(ctx) {
		if err != nil {
			return nil, err
		}

		// Filter for text/multimodal generation models
		// Include models that support generateContent (text and multimodal)
		// Exclude specialized output models (image gen, audio gen, TTS)
		supportsGenerateContent := false
		if m.SupportedActions != nil {
			for _, action := range m.SupportedActions {
				if action == "generateContent" {
					supportsGenerateContent = true
					break
				}
			}
		}

		if !supportsGenerateContent {
			continue
		}

		// Exclude output-focused models (image generation, audio generation, TTS)
		// Include: text-only and multimodal (text+image input) models
		// The "-image" suffix typically means image GENERATION output, not input
		name := m.Name
		if containsAny(name, []string{"-image", "-audio", "-video", "-tts", "imagen"}) {
			continue
		}

		models = append(models, name) // Name is like "models/gemini-pro" or "models/gemini-1.5-pro-vision"
	}
	return models, nil
}

// containsAny checks if string contains any of the substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
