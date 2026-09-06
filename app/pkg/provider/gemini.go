package provider

import (
	"HelAIx/pkg/gemini"
	"HelAIx/pkg/helix"
	"context"
	"fmt"
)

type GeminiService struct {
	apiKey string
	model  string
}

func NewGemini(apiKey, model string) (*GeminiService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API Key is missing")
	}
	return &GeminiService{apiKey: apiKey, model: model}, nil
}

func (s *GeminiService) Design(ctx context.Context, history []gemini.ChatMessage, variaxHardwareModel string) (*gemini.RigDescription, error) {
	client, err := gemini.NewClient(ctx, s.apiKey, s.model)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	return client.ChatSoundEngineer(ctx, history, variaxHardwareModel)
}

func (s *GeminiService) Build(ctx context.Context, rig gemini.RigDescription, presetName string, history []gemini.ChatMessage, hardware string, defaultExp int, variaxEnabled bool, variaxHardwareModel string) (*helix.Preset, error) {
	client, err := gemini.NewClient(ctx, s.apiKey, s.model)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	return client.ChatPresetEngineer(ctx, &rig, presetName, history, hardware, defaultExp, variaxEnabled, variaxHardwareModel)
}
