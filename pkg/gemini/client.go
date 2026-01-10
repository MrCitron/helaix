package gemini

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Client struct {
	model     *genai.GenerativeModel
	client    *genai.Client
	ModelName string
}

type ChatMessage struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

func NewClient(ctx context.Context, apiKey string, modelName string) (*Client, error) {
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	m := c.GenerativeModel(modelName)
	m.ResponseMIMEType = "application/json"

	return &Client{
		client:    c,
		model:     m,
		ModelName: modelName,
	}, nil
}

func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// ListModels returns a list of available generative models
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	iter := c.client.ListModels(ctx)
	var models []string
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		// Filter for generateContent supported models
		if m.SupportedGenerationMethods != nil {
			for _, method := range m.SupportedGenerationMethods {
				if method == "generateContent" {
					models = append(models, m.Name) // Name is like "models/gemini-pro"
					break
				}
			}
		}
	}
	return models, nil
}
