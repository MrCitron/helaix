package main

import (
	"HelAIx/pkg/config"
	"HelAIx/pkg/gemini"
	"HelAIx/pkg/helix"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"google.golang.org/genai"
)

const (
	maxLogBytes       int64 = 1 << 20
	maxLogEntryBytes        = 4 << 10
	logReadChunkBytes       = 4 << 10
)

// App struct
type App struct {
	ctx        context.Context
	config     *config.Manager
	logger     *log.Logger
	logFile    *os.File
	logPath    string
	logInitErr error
	logMu      sync.Mutex
}

// NewApp constructs the application and retains any logger setup error for the diagnostics UI.
func NewApp() *App {
	logger, logFile, logPath, logInitErr := newAppLogger()
	return &App{
		config:     config.NewManager(),
		logger:     logger,
		logFile:    logFile,
		logPath:    logPath,
		logInitErr: logInitErr,
	}
}

// newAppLogger opens the per-user log file and falls back to stderr if setup fails.
func newAppLogger() (*log.Logger, *os.File, string, error) {
	stderrLogger := log.New(os.Stderr, "helaix ", log.LstdFlags)
	configDir, err := os.UserConfigDir()
	if err != nil {
		return stderrLogger, nil, "", fmt.Errorf("resolve user config directory: %w", err)
	}

	logPath := filepath.Join(configDir, "helaix", "helaix.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
		return stderrLogger, nil, logPath, fmt.Errorf("create log directory: %w", err)
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return stderrLogger, nil, logPath, fmt.Errorf("open log file: %w", err)
	}
	return log.New(file, "helaix ", log.LstdFlags), file, logPath, nil
}

// logAIError appends a bounded diagnostic entry without recording prompts or generated content.
func (a *App) logAIError(operation string, model string, err error) {
	a.logMu.Lock()
	defer a.logMu.Unlock()

	if a.logger == nil || err == nil {
		return
	}

	var apiErr genai.APIError
	var entry string
	if errors.As(err, &apiErr) {
		entry = fmt.Sprintf("AI request failed operation=%s model=%s code=%d status=%q message=%q", operation, model, apiErr.Code, apiErr.Status, apiErr.Message)
	} else {
		entry = fmt.Sprintf("AI request failed operation=%s model=%s error=%q", operation, model, err.Error())
		if rawResponse := strings.Index(entry, ". Raw:"); rawResponse >= 0 {
			entry = entry[:rawResponse] + ". raw response omitted"
		}
	}
	if len(entry) > maxLogEntryBytes {
		entry = strings.ToValidUTF8(entry[:maxLogEntryBytes-15], "�") + " [truncated]"
	}

	if a.logFile != nil {
		info, statErr := a.logFile.Stat()
		if statErr != nil {
			log.Printf("helaix: inspect log file before append: %v", statErr)
			return
		}
		if info.Size()+int64(len(entry))+64 > maxLogBytes {
			if truncateErr := a.logFile.Truncate(0); truncateErr != nil {
				log.Printf("helaix: rotate log file before append: %v", truncateErr)
				return
			}
		}
	}
	a.logger.Print(entry)
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

// GxChatSoundEngineer calls the design agent and logs failures without storing prompt content.
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

	result, err := client.ChatSoundEngineer(a.ctx, history, cfg.VariaxHardwareModel, cfg.DefaultInstrument, cfg.VariaxEnabled)
	if err != nil {
		a.logAIError("sound_engineer", cfg.Model, err)
	}
	return result, err
}

// GxChatPresetEngineer builds or refines a preset and logs failures without storing prompt content.
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

	result, err := client.ChatPresetEngineer(a.ctx, &rig, presetName, history, cfg.HardwareTarget, cfg.DefaultExpPedal, cfg.VariaxEnabled, cfg.VariaxHardwareModel, cfg.DefaultInstrument)
	if err != nil {
		a.logAIError("preset_engineer", cfg.Model, err)
	}
	return result, err
}

// GxGetLogPath returns the local log path or its initialization error.
func (a *App) GxGetLogPath() (string, error) {
	if a.logInitErr != nil {
		return a.logPath, fmt.Errorf("application log is unavailable: %w", a.logInitErr)
	}
	return a.logPath, nil
}

// GxReadLog returns the most recent log lines for troubleshooting.
func (a *App) GxReadLog(maxLines int) (string, error) {
	if maxLines <= 0 || maxLines > 500 {
		maxLines = 200
	}

	a.logMu.Lock()
	defer a.logMu.Unlock()

	if a.logInitErr != nil {
		return "", fmt.Errorf("application log is unavailable: %w", a.logInitErr)
	}
	if a.logPath == "" {
		return "", fmt.Errorf("application log path is unavailable")
	}

	file, err := os.Open(a.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		return "", nil
	}

	start := info.Size() - maxLogBytes
	if start < 0 {
		start = 0
	}
	position := info.Size()
	lineBreaks := 0
	chunks := make([][]byte, 0)
	for position > start && lineBreaks <= maxLines {
		chunkSize := int64(logReadChunkBytes)
		if position-start < chunkSize {
			chunkSize = position - start
		}
		position -= chunkSize
		chunk := make([]byte, int(chunkSize))
		read, readErr := file.ReadAt(chunk, position)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return "", readErr
		}
		chunk = chunk[:read]
		lineBreaks += bytes.Count(chunk, []byte{'\n'})
		chunks = append(chunks, chunk)
	}

	data := make([]byte, 0, int(info.Size()-position))
	for i := len(chunks) - 1; i >= 0; i-- {
		data = append(data, chunks[i]...)
	}
	if position > 0 {
		previousByte := []byte{0}
		if _, readErr := file.ReadAt(previousByte, position-1); readErr != nil && !errors.Is(readErr, io.EOF) {
			return "", readErr
		}
		if previousByte[0] != '\n' {
			if firstLineBreak := bytes.IndexByte(data, '\n'); firstLineBreak >= 0 {
				data = data[firstLineBreak+1:]
			} else {
				data = nil
			}
		}
	}
	data = bytes.TrimRight(data, "\r\n")
	if len(data) == 0 {
		return "", nil
	}
	lines := bytes.Split(data, []byte{'\n'})
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return string(bytes.Join(lines, []byte{'\n'})), nil
}

// GxClearLog removes the local log contents while keeping the log file available.
func (a *App) GxClearLog() error {
	a.logMu.Lock()
	defer a.logMu.Unlock()

	if a.logInitErr != nil {
		return fmt.Errorf("application log is unavailable: %w", a.logInitErr)
	}
	if a.logPath == "" {
		return fmt.Errorf("application log path is unavailable")
	}
	return os.Truncate(a.logPath, 0)
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

// GxListModels lists generation-capable models and logs provider request failures.
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
		a.logAIError("list_models", modelName, err)
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

// GxTestConnection validates the API key by listing models and logs provider failures.
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
		a.logAIError("test_connection", modelName, err)
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
