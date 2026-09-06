package main

import (
	"HelAIx/pkg/codexcli"
	"HelAIx/pkg/config"
	"HelAIx/pkg/gemini"
	"HelAIx/pkg/helix"
	"HelAIx/pkg/provider"
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
	service, err := a.providerForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return service.Design(a.ctx, history, cfg.VariaxHardwareModel)
}

// GxChatPresetEngineer calls the Preset Engineer Agent with history and baseline rig
func (a *App) GxChatPresetEngineer(rig gemini.RigDescription, presetName string, history []gemini.ChatMessage) (*helix.Preset, error) {
	cfg := a.config.Get()
	service, err := a.providerForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return service.Build(a.ctx, rig, presetName, history, cfg.HardwareTarget, cfg.DefaultExpPedal, cfg.VariaxEnabled, cfg.VariaxHardwareModel)
}

func (a *App) providerForConfig(cfg config.AppConfig) (provider.Service, error) {
	switch config.NormalizeProvider(cfg.Provider) {
	case provider.Gemini:
		return provider.NewGemini(cfg.ApiKey, cfg.Model)
	case provider.CodexCLI:
		return codexcli.New(cfg.CodexCLIPath, cfg.Model)
	default:
		return nil, fmt.Errorf("unsupported AI provider %q", cfg.Provider)
	}
}

// GxGetCodexCLIStatus reports local Codex CLI installation, capabilities, and login method.
func (a *App) GxGetCodexCLIStatus(path string) provider.Status {
	return codexcli.Status(path)
}

// GxTestCodexCLI sends a minimal request through the locally logged-in Codex CLI.
func (a *App) GxTestCodexCLI(path, model string) (string, error) {
	return codexcli.TestConnection(a.ctx, path, model)
}

// GxSaveFile saves the preset to the disk and returns the full path
func (a *App) GxSaveFile(preset helix.Preset, filename string) (string, error) {
	if err := helix.ValidatePreset(preset); err != nil {
		return "", fmt.Errorf("refusing to export invalid preset: %w", err)
	}
	nameOnly, err := validateExportFilename(filename)
	if err != nil {
		return "", err
	}

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

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	ext := ".hlx"
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

func validateExportFilename(filename string) (string, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" || filepath.IsAbs(filename) || filepath.Base(filename) != filename || strings.Contains(filename, "\\") {
		return "", fmt.Errorf("filename must be a non-empty .hlx file name without path separators")
	}

	ext := ".hlx"
	if fileExt := filepath.Ext(filename); fileExt != "" && fileExt != ext {
		return "", fmt.Errorf("filename must use the %s extension", ext)
	}
	nameOnly := strings.TrimSuffix(filename, ext)
	if nameOnly == "" || nameOnly == "." || nameOnly == ".." {
		return "", fmt.Errorf("filename must contain a name")
	}
	return nameOnly, nil
}

// GxListModels returns the available models from the provider
func (a *App) GxListModels(apiKey string, modelName string) ([]string, error) {
	if apiKey == "" {
		return []string{}, nil
	}

	// Use fallback if no model name provided
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	// Create a temporary client just for listing
	client, err := gemini.NewClient(a.ctx, apiKey, modelName)
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
func (a *App) GxTestConnection(apiKey string, modelName string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("API Key is missing")
	}

	// Use fallback if no model name provided
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	// Create client and test connection
	client, err := gemini.NewClient(a.ctx, apiKey, modelName)
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
