package provider

import (
	"HelAIx/pkg/gemini"
	"HelAIx/pkg/helix"
	"context"
)

const (
	Gemini   = "gemini"
	CodexCLI = "codex_cli"
)

// Service keeps Wails independent from any specific LLM transport.
type Service interface {
	Design(ctx context.Context, history []gemini.ChatMessage, variaxHardwareModel string) (*gemini.RigDescription, error)
	Build(ctx context.Context, rig gemini.RigDescription, presetName string, history []gemini.ChatMessage, hardware string, defaultExp int, variaxEnabled bool, variaxHardwareModel string) (*helix.Preset, error)
}

type Status struct {
	Provider       string   `json:"provider"`
	Available      bool     `json:"available"`
	Configured     bool     `json:"configured"`
	Connected      bool     `json:"connected"`
	Models         []string `json:"models,omitempty"`
	Installed      bool     `json:"installed"`
	Executable     string   `json:"executable,omitempty"`
	Version        string   `json:"version,omitempty"`
	Authentication string   `json:"authentication,omitempty"`
	SupportsModel  bool     `json:"supports_model"`
	Message        string   `json:"message"`
}
