package codexcli

import (
	"HelAIx/pkg/gemini"
	"HelAIx/pkg/helix"
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

func TestBuildAppliesValidParameters(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_RESPONSE", validBuildResponseWithDriveParams())
	client, err := New(path, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	preset, err := client.Build(context.Background(), testBuildRig(), "TEST RIG", nil, "Helix Floor", 1, false, "Standard")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	drive := presetBlock(t, preset, "dsp0", "block0")
	if got := drive["Gain"]; got != 0.2 {
		t.Fatalf("Drive Gain = %v, want 0.2", got)
	}
	if got := drive["Tone"]; got != 0.4 {
		t.Fatalf("Drive Tone = %v, want 0.4", got)
	}
}

func TestBuildSerializesMultipleBlockParameters(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_RESPONSE", validBuildResponseWithMultipleBlockParams())
	client, err := New(path, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	preset, err := client.Build(context.Background(), testBuildRig(), "TEST RIG", nil, "Helix Floor", 1, false, "Standard")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	encoded, err := json.Marshal(preset)
	if err != nil {
		t.Fatalf("marshal preset: %v", err)
	}
	var decoded helix.Preset
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal preset: %v", err)
	}

	assertBlockParameters(t, presetBlock(t, &decoded, "dsp0", "block0"), map[string]interface{}{
		"Drive":    0.33,
		"Bass Cut": true,
		"Voltage":  false,
	})
	assertBlockParameters(t, presetBlock(t, &decoded, "dsp0", "block1"), map[string]interface{}{
		"Drive":  0.35,
		"Bass":   0.45,
		"Master": 0.75,
	})
	assertBlockParameters(t, presetBlock(t, &decoded, "dsp0", "block2"), map[string]interface{}{
		"Distance": 3.0,
		"HighCut":  8000.0,
		"LowCut":   80.0,
		"Mic":      4.0,
		"Position": 0.5,
	})
}

func TestBuildRejectsOutOfRangeParameters(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_RESPONSE", outOfRangeBuildResponse())
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

func TestBuildRetriesOutOfRangeParameters(t *testing.T) {
	path := fakeCodex(t)
	t.Setenv("FAKE_RESPONSE", outOfRangeBuildResponse())
	t.Setenv("FAKE_RETRY_RESPONSE", validBuildResponseWithDriveParams())
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
	assertCodexCalls(t, callsPath, 2)
	if got := presetBlock(t, preset, "dsp0", "block0")["Gain"]; got != 0.2 {
		t.Fatalf("corrected Drive Gain = %v, want 0.2", got)
	}
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("read retry prompt: %v", err)
	}
	if !strings.Contains(string(prompt), "outside the allowed range") {
		t.Fatalf("retry prompt does not explain the out-of-range parameter: %s", prompt)
	}
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
	if !strings.Contains(prompt, "optional params object") {
		t.Fatal("presetPrompt() does not instruct Codex to dial model parameters")
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

func TestBlocksSchemaAllowsPrimitiveParameters(t *testing.T) {
	var schema map[string]interface{}
	if err := json.Unmarshal(blocksSchema(), &schema); err != nil {
		t.Fatalf("unmarshal blocks schema: %v", err)
	}
	properties := schema["properties"].(map[string]interface{})
	blocks := properties["blocks"].(map[string]interface{})
	items := blocks["items"].(map[string]interface{})
	blockProperties := items["properties"].(map[string]interface{})
	params, exists := blockProperties["params"].(map[string]interface{})
	if !exists || params["type"] != "object" {
		t.Fatalf("params schema = %v, want object", params)
	}
	values, exists := params["additionalProperties"].(map[string]interface{})
	if !exists {
		t.Fatalf("params schema = %v, want primitive value schema", params)
	}
	allowed := values["type"].([]interface{})
	if len(allowed) != 3 || allowed[0] != "number" || allowed[1] != "string" || allowed[2] != "boolean" {
		t.Fatalf("parameter types = %v, want number, string, boolean", allowed)
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

func validBuildResponseWithDriveParams() string {
	return `{"blocks":[{"name":"Drive","model_name":"Scream 808","path":0,"params":{"Gain":0.2,"Tone":0.4}},{"name":"Amp","model_name":"US Deluxe Nrm","path":0,"params":{}},{"name":"Cab","model_name":"1x12 US Deluxe","path":0,"params":{}}]}`
}

func validBuildResponseWithMultipleBlockParams() string {
	return `{"blocks":[{"name":"Drive","model_name":"Prize Drive","path":0,"params":{"Drive":0.33,"Bass Cut":true,"Voltage":false}},{"name":"Amp","model_name":"US Deluxe Nrm","path":0,"params":{"Drive":0.35,"Bass":0.45,"Master":0.75}},{"name":"Cab","model_name":"1x12 US Deluxe","path":0,"params":{"Distance":3,"HighCut":8000,"LowCut":80,"Mic":4,"Position":0.5}}]}`
}

func outOfRangeBuildResponse() string {
	return `{"blocks":[{"name":"Drive","model_name":"Scream 808","path":0,"params":{"Gain":1.1}},{"name":"Amp","model_name":"US Deluxe Nrm","path":0,"params":{}},{"name":"Cab","model_name":"1x12 US Deluxe","path":0,"params":{}}]}`
}

func presetBlock(t *testing.T, preset *helix.Preset, dspName, blockName string) map[string]interface{} {
	t.Helper()
	data, ok := (*preset)["data"].(map[string]interface{})
	if !ok {
		t.Fatal("preset has no data")
	}
	tone, ok := data["tone"].(map[string]interface{})
	if !ok {
		t.Fatal("preset has no tone")
	}
	dsp, ok := tone[dspName].(map[string]interface{})
	if !ok {
		t.Fatalf("preset has no %s", dspName)
	}
	block, ok := dsp[blockName].(map[string]interface{})
	if !ok {
		t.Fatalf("preset has no %s in %s", blockName, dspName)
	}
	return block
}

func assertBlockParameters(t *testing.T, block map[string]interface{}, want map[string]interface{}) {
	t.Helper()
	for name, value := range want {
		if got := block[name]; got != value {
			t.Errorf("parameter %q = %v, want %v", name, got, value)
		}
	}
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
