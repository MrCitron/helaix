package main

import (
	"HelAIx/pkg/config"
	"HelAIx/pkg/gemini"
	"HelAIx/pkg/helix"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	config *config.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		config: config.NewManager(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GxGetConfig returns the current configuration
func (a *App) GxGetConfig() config.AppConfig {
	return a.config.Get()
}

// GxSaveConfig saves the configuration
func (a *App) GxSaveConfig(cfg config.AppConfig) string {
	err := a.config.Save(cfg)
	if err != nil {
		return fmt.Sprintf("Error saving config: %s", err.Error())
	}
	return ""
}

// GxChatSoundEngineer calls the Sound Engineer Agent with history
func (a *App) GxChatSoundEngineer(history []gemini.ChatMessage) (*gemini.RigDescription, error) {
	cfg := a.config.Get()
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("API Key is missing")
	}

	client, err := gemini.NewClient(a.ctx, cfg.ApiKey, cfg.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %v", err)
	}
	defer client.Close()

	return client.ChatSoundEngineer(a.ctx, history, cfg.VariaxHardwareModel)
}

// GxChatPresetEngineer calls the Preset Engineer Agent with history and baseline rig
func (a *App) GxChatPresetEngineer(rig gemini.RigDescription, presetName string, history []gemini.ChatMessage) (*helix.Preset, error) {
	cfg := a.config.Get()
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("API Key is missing")
	}

	client, err := gemini.NewClient(a.ctx, cfg.ApiKey, cfg.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %v", err)
	}
	defer client.Close()

	return client.ChatPresetEngineer(a.ctx, &rig, presetName, history, cfg.HardwareTarget, cfg.DefaultExpPedal, cfg.VariaxEnabled, cfg.VariaxHardwareModel)
}

// GxSaveFile saves the preset to the disk and returns the full path
func (a *App) GxSaveFile(preset helix.Preset, filename string) (string, error) {
	cfg := a.config.Get()
	// Use default path if absolute path not provided (simplified)
	baseDir := cfg.OutputPath
	if !filepath.IsAbs(baseDir) {
		// Use HOME for macOS/Linux, USERPROFILE for Windows
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE")
		}
		baseDir = filepath.Join(homeDir, "Documents", "helaix")
	}

	// Ensure dir
	os.MkdirAll(baseDir, 0755)

	// Clean filename and ensure extension
	ext := ".hlx"
	nameOnly := strings.TrimSuffix(filename, ext)

	fullPath := filepath.Join(baseDir, nameOnly+ext)

	// Incremental logic
	if cfg.IncrementalSave {
		counter := 1
		for {
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				break
			}
			fullPath = filepath.Join(baseDir, fmt.Sprintf("%s_%d%s", nameOnly, counter, ext))
			counter++
		}
	}

	data, err := json.MarshalIndent(preset, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(fullPath, data, 0644)
	return fullPath, err
}

// GxListModels returns the available models from the provider
func (a *App) GxListModels() ([]string, error) {
	cfg := a.config.Get()
	if cfg.ApiKey == "" {
		return []string{}, nil
	}

	// Get the first available model name from config or use fallback
	modelName := cfg.Model
	if modelName == "" {
		modelName = "gemini-2.5-flash" // Fallback only if no model in config
	}

	// Create a temporary client just for listing
	client, err := gemini.NewClient(a.ctx, cfg.ApiKey, modelName)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	models, err := client.ListModels(a.ctx)
	if err != nil {
		return nil, err
	}

	// Clean up model names (remove "models/" prefix)
	var cleanModels []string
	for _, m := range models {
		if len(m) > 7 && m[:7] == "models/" {
			cleanModels = append(cleanModels, m[7:])
		} else {
			cleanModels = append(cleanModels, m)
		}
	}
	return cleanModels, nil
}

// GxSelectFolder opens a directory dialog and returns the selected path
func (a *App) GxSelectFolder(initialDir string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Select Export Folder",
		DefaultDirectory: initialDir,
	})
}

// GxGetDefaultOutputPath returns the default Documents/helaix path
func (a *App) GxGetDefaultOutputPath() string {
	// Use HOME for macOS/Linux, USERPROFILE for Windows
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = os.Getenv("USERPROFILE") // Windows fallback
	}
	return filepath.Join(homeDir, "Documents", "helaix")
}

// GxTestConnection validates the API key by listing models
func (a *App) GxTestConnection() (string, error) {
	cfg := a.config.Get()
	if cfg.ApiKey == "" {
		return "", fmt.Errorf("API Key is missing")
	}

	// Get the model name from config or use fallback
	modelName := cfg.Model
	if modelName == "" {
		modelName = "gemini-2.5-flash" // Fallback only if no model in config
	}

	// Create client and test connection
	client, err := gemini.NewClient(a.ctx, cfg.ApiKey, modelName)
	if err != nil {
		return "", err
	}
	defer client.Close()

	models, err := client.ListModels(a.ctx)
	if err != nil {
		return "", err
	}

	// Return success with count of available models
	return fmt.Sprintf("Connection successful! Found %d available models.", len(models)), nil
}

// GxOpenPath opens the given path (file or folder) using the system's default application
func (a *App) GxOpenPath(path string) {
	if path == "" {
		return
	}
	runtime.BrowserOpenURL(a.ctx, path)
}

// GxOpenFolderOfFile extracts the directory from a file path and opens it
func (a *App) GxOpenFolderOfFile(filePath string) {
	if filePath == "" {
		return
	}
	dir := filepath.Dir(filePath)
	runtime.BrowserOpenURL(a.ctx, dir)
}
