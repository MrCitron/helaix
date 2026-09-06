package codexcli

import (
	"HelAIx/pkg/gemini"
	"HelAIx/pkg/helix"
	"HelAIx/pkg/provider"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrCLIUnavailable   = errors.New("Codex CLI is not installed or executable")
	ErrNotAuthenticated = errors.New("Codex CLI is not signed in with ChatGPT")
	ErrAPIKeyAuth       = errors.New("Codex CLI is using API key authentication, not ChatGPT subscription")
	ErrModelUnsupported = errors.New("this Codex CLI version does not report --model support")
	ErrQuota            = errors.New("Codex subscription quota is exhausted")
	ErrTimeout          = errors.New("Codex CLI request timed out")
	ErrInvalidResponse  = errors.New("Codex CLI returned an invalid structured response")
	requestTimeout      = 90 * time.Second
	statusTimeout       = 5 * time.Second
)

type Client struct {
	path          string
	model         string
	supportsModel bool
}

func New(path, model string) (*Client, error) {
	status := Status(path)
	if !status.Installed {
		return nil, ErrCLIUnavailable
	}
	if status.Authentication == "api" {
		return nil, ErrAPIKeyAuth
	}
	if !status.Connected || status.Authentication != "chatgpt" {
		return nil, ErrNotAuthenticated
	}
	if model != "" && !status.SupportsModel {
		return nil, ErrModelUnsupported
	}
	return &Client{path: status.Executable, model: model, supportsModel: status.SupportsModel}, nil
}

// Status only asks Codex for version, capabilities, login method, and its local model catalog. It never reads credential files.
func Status(configuredPath string) provider.Status {
	status := provider.Status{Provider: provider.CodexCLI}
	path, err := resolveExecutable(configuredPath)
	if err != nil {
		status.Message = "Codex CLI was not found. Install it, or set its executable path in Settings."
		return status
	}
	status.Available = true
	status.Installed = true
	status.Executable = path

	if output, err := runStatusCommand(path, "--version"); err == nil {
		status.Version = strings.TrimSpace(string(output))
	}
	if output, err := runStatusCommand(path, "exec", "--help"); err == nil {
		status.SupportsModel = strings.Contains(string(output), "--model")
	}

	output, err := runStatusCommand(path, "login", "status")
	if err != nil {
		status.Message = "Codex CLI is installed but not signed in. Run `codex login` in Terminal and choose ChatGPT sign-in."
		return status
	}
	login := strings.ToLower(string(output))
	switch {
	case strings.Contains(login, "chatgpt"):
		status.Configured = true
		status.Connected = true
		status.Authentication = "chatgpt"
		status.Message = "Codex CLI is signed in with ChatGPT. Requests use your Codex plan limits."
		if output, err := runStatusCommand(path, "debug", "models"); err == nil {
			status.Models = visibleModels(output)
		}
	case strings.Contains(login, "api key") || strings.Contains(login, "api-key"):
		status.Authentication = "api"
		status.Message = "Codex CLI is signed in with an API key. Run `codex logout`, then `codex login` and choose ChatGPT."
	default:
		status.Message = "Codex CLI login method could not be verified. Run `codex login status` in Terminal."
	}
	return status
}

func visibleModels(output []byte) []string {
	var catalog struct {
		Models []struct {
			Slug       string `json:"slug"`
			Visibility string `json:"visibility"`
		} `json:"models"`
	}
	if err := json.Unmarshal(output, &catalog); err != nil {
		return nil
	}
	models := make([]string, 0, len(catalog.Models))
	for _, model := range catalog.Models {
		if model.Slug != "" && model.Visibility == "list" {
			models = append(models, model.Slug)
		}
	}
	return models
}

func (c *Client) Design(ctx context.Context, history []gemini.ChatMessage, hardwareModel string) (*gemini.RigDescription, error) {
	prompt := soundPrompt(history, hardwareModel)
	output, err := c.run(ctx, prompt, rigSchema())
	if err != nil {
		return nil, err
	}
	rig, err := gemini.ParseRigDescriptionJSON(output)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	return rig, nil
}

func (c *Client) Build(ctx context.Context, rig gemini.RigDescription, presetName string, history []gemini.ChatMessage, hardware string, defaultExp int, variaxEnabled bool, variaxHardwareModel string) (*helix.Preset, error) {
	prompt, err := presetPrompt(rig, history, hardware)
	if err != nil {
		return nil, err
	}
	output, err := c.run(ctx, prompt, blocksSchema())
	if err != nil {
		return nil, err
	}
	preset, err := gemini.BuildPresetFromJSON(output, &rig, presetName, hardware, defaultExp, variaxEnabled, variaxHardwareModel)
	if err == nil {
		return preset, nil
	}
	retryPrompt := fmt.Sprintf("%s\n\nBEGIN REJECTED JSON\n%s\nEND REJECTED JSON\n\n%s", prompt, output, gemini.BuilderCorrectionInstruction(err))
	retryOutput, retryErr := c.run(ctx, retryPrompt, blocksSchema())
	if retryErr != nil {
		return nil, retryErr
	}
	preset, retryErr = gemini.BuildPresetFromJSON(retryOutput, &rig, presetName, hardware, defaultExp, variaxEnabled, variaxHardwareModel)
	if retryErr != nil {
		return nil, fmt.Errorf("%w: response remained invalid after retry: %v", ErrInvalidResponse, retryErr)
	}
	return preset, nil
}

// TestConnection deliberately makes a minimal Codex request and therefore consumes plan quota.
func TestConnection(ctx context.Context, path, model string) (string, error) {
	client, err := New(path, model)
	if err != nil {
		return "", err
	}
	result, err := client.run(ctx, "Return only JSON confirming that Codex CLI is available to HelAIx.", connectionSchema())
	if err != nil {
		return "", err
	}
	var response struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal([]byte(result), &response); err != nil || !response.OK {
		return "", ErrInvalidResponse
	}
	return "Codex CLI connection succeeded using your ChatGPT subscription.", nil
}

func (c *Client) run(parent context.Context, prompt string, schema []byte) (string, error) {
	return c.runAttempt(parent, prompt, schema, true)
}

func (c *Client) runAttempt(parent context.Context, prompt string, schema []byte, retryAllowed bool) (string, error) {
	ctx, cancel := context.WithTimeout(parent, requestTimeout)
	defer cancel()

	workDir, err := os.MkdirTemp("", "helaix-codex-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workDir)
	if err := os.Chmod(workDir, 0700); err != nil {
		return "", err
	}
	if output, err := exec.Command("git", "init", "--quiet", workDir).CombinedOutput(); err != nil {
		return "", fmt.Errorf("prepare isolated Codex workspace: %s", strings.TrimSpace(string(output)))
	}

	schemaPath := filepath.Join(workDir, "response-schema.json")
	outputPath := filepath.Join(workDir, "response.json")
	if err := os.WriteFile(schemaPath, schema, 0600); err != nil {
		return "", err
	}
	args := []string{"exec", "--ephemeral", "--sandbox", "read-only", "--json", "--output-schema", schemaPath, "-o", outputPath}
	if c.model != "" && c.supportsModel {
		args = append(args, "--model", c.model)
	}
	args = append(args, "-")
	cmd := exec.CommandContext(ctx, c.path, args...)
	cmd.Dir = workDir
	cmd.Env = sanitizedEnvironment(os.Environ())
	cmd.Stdin = strings.NewReader(prompt)
	combined, err := cmd.CombinedOutput()
	if err != nil {
		if retryAllowed && isTransientOutput(combined) && ctx.Err() == nil {
			select {
			case <-parent.Done():
				return "", parent.Err()
			case <-time.After(250 * time.Millisecond):
			}
			return c.runAttempt(parent, prompt, schema, false)
		}
		return "", classifyProcessError(ctx, combined, err)
	}
	response, err := os.ReadFile(outputPath)
	if err != nil || len(bytes.TrimSpace(response)) == 0 {
		return "", ErrInvalidResponse
	}
	return string(response), nil
}

func isTransientOutput(output []byte) bool {
	text := strings.ToLower(string(output))
	return strings.Contains(text, "temporarily unavailable") ||
		strings.Contains(text, "connection reset") ||
		strings.Contains(text, "network error") ||
		strings.Contains(text, "status 503")
}

func resolveExecutable(configuredPath string) (string, error) {
	candidates := []string{configuredPath}
	if path, err := exec.LookPath("codex"); err == nil {
		candidates = append(candidates, path)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, ".local", "bin", "codex"),
			filepath.Join(home, ".npm-global", "bin", "codex"),
			filepath.Join(home, ".volta", "bin", "codex"),
		)
	}
	candidates = append(candidates, "/opt/homebrew/bin/codex", "/usr/local/bin/codex")
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return candidate, nil
		}
	}
	return "", ErrCLIUnavailable
}

func runStatusCommand(path string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), statusTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = sanitizedEnvironment(os.Environ())
	return cmd.CombinedOutput()
}

func sanitizedEnvironment(environment []string) []string {
	blocked := map[string]bool{
		"OPENAI_API_KEY":     true,
		"CODEX_API_KEY":      true,
		"CODEX_ACCESS_TOKEN": true,
		"OPENAI_BASE_URL":    true,
	}
	var clean []string
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if !blocked[name] {
			clean = append(clean, entry)
		}
	}
	return clean
}

func classifyProcessError(ctx context.Context, output []byte, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ErrTimeout
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return context.Canceled
	}
	text := strings.ToLower(string(output))
	switch {
	case strings.Contains(text, "quota") || strings.Contains(text, "rate limit"):
		return ErrQuota
	case strings.Contains(text, "not logged") || strings.Contains(text, "sign in") || strings.Contains(text, "login"):
		return ErrNotAuthenticated
	case strings.Contains(text, "model") && strings.Contains(text, "not available"):
		return ErrModelUnsupported
	default:
		return fmt.Errorf("Codex CLI request failed: %s", safeProcessSummary(output, err))
	}
}

func safeProcessSummary(output []byte, err error) string {
	summary := strings.Join(strings.Fields(string(output)), " ")
	if summary == "" {
		summary = err.Error()
	}
	if len(summary) > 400 {
		summary = summary[:400] + "..."
	}
	return summary
}
