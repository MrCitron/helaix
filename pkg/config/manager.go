package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type AppConfig struct {
	ApiKey          string `json:"api_key"`
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	OutputPath      string `json:"output_path"`
	HardwareTarget  string `json:"hardware_target"`
	DeleteNoConfirm bool   `json:"delete_no_confirm"`
	IncrementalSave bool   `json:"incremental_save"`
}

type Manager struct {
	mu         sync.RWMutex
	config     AppConfig
	configPath string
}

func NewManager() *Manager {
	configDir, _ := os.UserConfigDir()
	configPath := filepath.Join(configDir, "helaix", "settings.json")

	defaultOutPath := filepath.Join(os.Getenv("USERPROFILE"), "Documents", "helaix")

	m := &Manager{
		configPath: configPath,
		config: AppConfig{
			Provider:       "Google",
			Model:          "gemini-1.5-flash",
			OutputPath:     defaultOutPath,
			HardwareTarget: "Helix Floor",
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
