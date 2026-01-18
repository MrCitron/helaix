package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type AppConfig struct {
	ApiKey              string `json:"api_key"`
	Provider            string `json:"provider"` // "Google" = Gemini API (ai.google.dev), "Vertex" = Vertex AI
	Model               string `json:"model"`    // e.g., "gemini-2.5-flash", "gemini-3-flash-preview"
	OutputPath          string `json:"output_path"`
	HardwareTarget      string `json:"hardware_target"`
	DeleteNoConfirm     bool   `json:"delete_no_confirm"`
	IncrementalSave     bool   `json:"incremental_save"`
	DefaultExpPedal     int    `json:"default_exp_pedal"`     // 0 = None, 1 = Exp 1, 2 = Exp 2, 3 = Exp 3
	VariaxEnabled       bool   `json:"variax_enabled"`        // Whether to control Variax
	VariaxHardwareModel string `json:"variax_hardware_model"` // JTV, Standard, Shuriken
}

type Manager struct {
	mu         sync.RWMutex
	config     AppConfig
	configPath string
}

func NewManager() *Manager {
	configDir, _ := os.UserConfigDir()
	configPath := filepath.Join(configDir, "helaix", "settings.json")

	// Use HOME for macOS/Linux, USERPROFILE for Windows
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = os.Getenv("USERPROFILE") // Windows fallback
	}
	defaultOutPath := filepath.Join(homeDir, "Documents", "helaix")

	m := &Manager{
		configPath: configPath,
		config: AppConfig{
			Provider:            "Google",           // Gemini API (not Vertex AI)
			Model:               "gemini-2.5-flash", // Updated to current stable model
			OutputPath:          defaultOutPath,
			HardwareTarget:      "Helix Floor",
			DefaultExpPedal:     1, // Default to Exp 1
			VariaxEnabled:       false,
			VariaxHardwareModel: "Standard",
		},
	}
	m.Load()
	return m
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.config)
}

func (m *Manager) Save(cfg AppConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = cfg

	// Ensure dir exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.configPath, data, 0644)
}

func (m *Manager) Get() AppConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}
