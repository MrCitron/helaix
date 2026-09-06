package codexcli

import (
	"HelAIx/pkg/gemini"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClientDesignUsesReadOnlyStructuredExecution(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_RESPONSE", `{"suggested_name":"TEST RIG","explanation":"A test rig.","guitar_model":"Fender Stratocaster","tuning":"Standard","chain":[{"type":"amp","name":"Amp","description":"Test amp","settings":"Clean"},{"type":"cab","name":"Cab","description":"Test cab","settings":"2x12"}]}`)
	argsPath := filepath.Join(t.TempDir(), "args")
	t.Setenv("FAKE_ARGS_PATH", argsPath)

	client, err := New(path, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := client.Design(context.Background(), nil, "Standard"); err != nil {
		t.Fatalf("Design() error = %v", err)
	}
	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read args: %v", err)
	}
	for _, want := range []string{"--ephemeral", "--sandbox", "read-only", "--json", "--output-schema"} {
		if !strings.Contains(string(args), want) {
			t.Fatalf("Codex args %q do not include %q", args, want)
		}
	}
}

func TestClientRejectsInvalidStructuredResponse(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_RESPONSE", `{}`)
	client, err := New(path, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = client.Design(context.Background(), nil, "Standard")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("Design() error = %v, want ErrInvalidResponse", err)
	}
}

func TestBuildRejectsUnknownHelixModel(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_RESPONSE", invalidBuildResponse())
	callsPath := filepath.Join(t.TempDir(), "calls")
	t.Setenv("FAKE_CALLS_PATH", callsPath)
	client, err := New(path, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = client.Build(context.Background(), testBuildRig(), "TEST RIG", nil, "Helix Floor", 1, false, "Standard")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("Build() error = %v, want ErrInvalidResponse", err)
	}
	assertCodexCalls(t, callsPath, 2)
}

func TestBuildRetriesInvalidHelixResponse(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_RESPONSE", invalidBuildResponse())
	t.Setenv("FAKE_RETRY_RESPONSE", validBuildResponse())
	callsPath := filepath.Join(t.TempDir(), "calls")
	t.Setenv("FAKE_CALLS_PATH", callsPath)
	promptPath := filepath.Join(t.TempDir(), "prompt")
	t.Setenv("FAKE_PROMPT_PATH", promptPath)
	client, err := New(path, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	preset, err := client.Build(context.Background(), testBuildRig(), "TEST RIG", nil, "Helix Floor", 1, false, "Standard")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if preset == nil {
		t.Fatal("Build() returned a nil preset")
	}
	assertCodexCalls(t, callsPath, 2)
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("read retry prompt: %v", err)
	}
	for _, want := range []string{"BEGIN REJECTED JSON", "Invented Helix Model", "complete replacement JSON object"} {
		if !strings.Contains(string(prompt), want) {
			t.Fatalf("retry prompt does not include %q", want)
		}
	}
}

func TestBuildDoesNotRetryValidHelixResponse(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_RESPONSE", validBuildResponse())
	callsPath := filepath.Join(t.TempDir(), "calls")
	t.Setenv("FAKE_CALLS_PATH", callsPath)
	client, err := New(path, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := client.Build(context.Background(), testBuildRig(), "TEST RIG", nil, "Helix Floor", 1, false, "Standard"); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	assertCodexCalls(t, callsPath, 1)
}

func TestPresetPromptListsOnlyComponentCandidates(t *testing.T) {
	prompt, err := presetPrompt(gemini.RigDescription{
		Chain: []gemini.RigComponent{{Type: "amp", Name: "Amp"}},
	}, nil, "Helix Floor")
	if err != nil {
		t.Fatalf("presetPrompt() error = %v", err)
	}
	if !strings.Contains(prompt, "Amp [amp, ranked by tone intent]:") {
		t.Fatalf("presetPrompt() does not include typed amp candidates")
	}
	if strings.Contains(prompt, "Scream 808") {
		t.Fatal("presetPrompt() included a pedal model for an amp-only rig")
	}
	if !strings.Contains(prompt, "RECOMMENDED DSP-SAFE STARTING PLAN:") {
		t.Fatal("presetPrompt() does not include the DSP-safe recommendation")
	}
	if !strings.Contains(prompt, "complete, validated combination") {
		t.Fatal("presetPrompt() does not require the planner tuple to be preserved")
	}
}

func TestClientReturnsTimeoutAndCancellation(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_MODE", "sleep")
	client, err := New(path, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	oldTimeout := requestTimeout
	requestTimeout = 20 * time.Millisecond
	t.Cleanup(func() { requestTimeout = oldTimeout })
	_, err = client.Design(context.Background(), nil, "Standard")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Design() timeout error = %v, want ErrTimeout", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Design(ctx, nil, "Standard")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Design() cancel error = %v, want context.Canceled", err)
	}
}

func TestNewRejectsAPIKeyAuthenticationAndSanitizesChildEnvironment(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_LOGIN", "missing")
	if _, err := New(path, ""); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("New() error = %v, want ErrNotAuthenticated", err)
	}

	t.Setenv("FAKE_LOGIN", "api")
	if _, err := New(path, ""); !errors.Is(err, ErrAPIKeyAuth) {
		t.Fatalf("New() error = %v, want ErrAPIKeyAuth", err)
	}

	t.Setenv("FAKE_LOGIN", "chatgpt")
	t.Setenv("OPENAI_API_KEY", "must-not-reach-codex")
	envPath := filepath.Join(t.TempDir(), "env")
	t.Setenv("FAKE_ENV_PATH", envPath)
	client, err := New(path, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Setenv("FAKE_RESPONSE", `{"ok":true}`)
	if _, err := client.run(context.Background(), "test", connectionSchema()); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	environment, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("read child environment: %v", err)
	}
	if strings.Contains(string(environment), "OPENAI_API_KEY=") {
		t.Fatal("Codex child process inherited OPENAI_API_KEY")
	}
}

func TestSchemasAreValidJSON(t *testing.T) {
	for _, schema := range [][]byte{rigSchema(), blocksSchema(), connectionSchema()} {
		if !json.Valid(schema) {
			t.Fatalf("schema is invalid JSON: %s", schema)
		}
	}
}

func TestTransientRetryClassificationExcludesQuota(t *testing.T) {
	if !isTransientOutput([]byte("network error: temporarily unavailable")) {
		t.Fatal("transient network failure was not classified as retryable")
	}
	if isTransientOutput([]byte("subscription quota exhausted")) {
		t.Fatal("quota exhaustion must not be retried")
	}
}

func TestStatusListsVisibleModelsFromCatalog(t *testing.T) {
	status := Status(fakeCodex(t))
	want := []string{"gpt-6-astra", "gpt-5.6-terra"}
	if !status.Connected || !status.SupportsModel {
		t.Fatalf("Status() = %+v, want connected CLI with model support", status)
	}
	if len(status.Models) != len(want) {
		t.Fatalf("Status().Models = %v, want %v", status.Models, want)
	}
	for i, model := range want {
		if status.Models[i] != model {
			t.Fatalf("Status().Models = %v, want %v", status.Models, want)
		}
	}
}

func fakeCodex(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "codex")
	script := `#!/bin/sh
if [ "$1" = "--version" ]; then echo "codex 1.0.0"; exit 0; fi
if [ "$1" = "login" ]; then
  if [ "$FAKE_LOGIN" = "missing" ]; then exit 1; fi
  if [ "$FAKE_LOGIN" = "api" ]; then echo "Logged in with API key"; else echo "Logged in with ChatGPT"; fi
  exit 0
fi
if [ "$1" = "exec" ] && [ "$2" = "--help" ]; then echo "--model --sandbox --json --output-schema"; exit 0; fi
if [ "$1" = "debug" ] && [ "$2" = "models" ]; then
  echo '{"models":[{"slug":"gpt-6-astra","visibility":"list"},{"slug":"gpt-reserve","visibility":"hide"},{"slug":"gpt-5.6-terra","visibility":"list"}]}'
  exit 0
fi
if [ "$1" = "exec" ]; then
  if [ "$FAKE_MODE" = "sleep" ]; then sleep 2; fi
  if [ -n "$FAKE_ARGS_PATH" ]; then printf '%s\n' "$@" > "$FAKE_ARGS_PATH"; fi
  if [ -n "$FAKE_ENV_PATH" ]; then env > "$FAKE_ENV_PATH"; fi
  if [ -n "$FAKE_PROMPT_PATH" ]; then cat > "$FAKE_PROMPT_PATH"; fi
  previous=""
  for arg in "$@"; do
    if [ "$previous" = "-o" ]; then output="$arg"; break; fi
    previous="$arg"
  done
  response="$FAKE_RESPONSE"
  if [ -n "$FAKE_CALLS_PATH" ]; then
    if [ -s "$FAKE_CALLS_PATH" ] && [ -n "$FAKE_RETRY_RESPONSE" ]; then response="$FAKE_RETRY_RESPONSE"; fi
    printf x >> "$FAKE_CALLS_PATH"
  fi
  printf '%s' "$response" > "$output"
  exit 0
fi
exit 1
`
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		t.Fatalf("write fake Codex: %v", err)
	}
	return path
}

func testBuildRig() gemini.RigDescription {
	return gemini.RigDescription{
		SuggestedName: "TEST RIG",
		Explanation:   "A test rig.",
		GuitarModel:   "Fender Stratocaster",
		Tuning:        "Standard",
		Chain: []gemini.RigComponent{
			{Type: "pedal", Name: "Drive", Description: "Overdrive", Settings: "Low gain"},
			{Type: "amp", Name: "Amp", Description: "Clean amp", Settings: "Clean"},
			{Type: "cab", Name: "Cab", Description: "1x12 cabinet", Settings: "1x12"},
		},
	}
}

func invalidBuildResponse() string {
	return `{"blocks":[{"name":"Drive","model_name":"Invented Helix Model","path":0,"params":{}},{"name":"Amp","model_name":"US Deluxe Nrm","path":0,"params":{}},{"name":"Cab","model_name":"1x12 US Deluxe","path":0,"params":{}}]}`
}

func validBuildResponse() string {
	return `{"blocks":[{"name":"Drive","model_name":"Scream 808","path":0,"params":{}},{"name":"Amp","model_name":"US Deluxe Nrm","path":0,"params":{}},{"name":"Cab","model_name":"1x12 US Deluxe","path":0,"params":{}}]}`
}

func assertCodexCalls(t *testing.T, path string, want int) {
	t.Helper()
	calls, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read call count: %v", err)
	}
	if len(calls) != want {
		t.Fatalf("Codex calls = %d, want %d", len(calls), want)
	}
}
