package codexcli

import (
	"HelAIx/pkg/gemini"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestLiveChatGPTSubscription is intentionally opt-in because it consumes Codex plan quota.
func TestLiveChatGPTSubscription(t *testing.T) {
	if os.Getenv("HELAIX_RUN_CODEX_LIVE") != "1" {
		t.Skip("set HELAIX_RUN_CODEX_LIVE=1 to exercise the local ChatGPT Codex session")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	if _, err := TestConnection(ctx, "", ""); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}

	client, err := New("", "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	history := []gemini.ChatMessage{{
		Role:    "user",
		Content: "Create a clean, low-gain ambient rhythm rig with one amp, one cab, short delay, and reverb.",
	}}
	rig, err := client.Design(ctx, history, "Standard")
	if err != nil {
		t.Fatalf("Design() error = %v", err)
	}
	if _, err := client.Build(ctx, *rig, rig.SuggestedName, history, "Helix Floor", 1, false, "Standard"); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

// TestLiveChatGPTSubscriptionAcceptsDynamicParams verifies Codex accepts the dynamic params schema.
func TestLiveChatGPTSubscriptionAcceptsDynamicParams(t *testing.T) {
	if os.Getenv("HELAIX_RUN_CODEX_LIVE") != "1" {
		t.Skip("set HELAIX_RUN_CODEX_LIVE=1 to exercise the local ChatGPT Codex session")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	client, err := New("", "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	output, err := client.run(ctx, `Return exactly one block with name "Drive", model_name "Scream 808", path 0, and params {"Gain": 0.2, "Tone": 0.4}.`, blocksSchema())
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	var response struct {
		Blocks []struct {
			Name      string                 `json:"name"`
			ModelName string                 `json:"model_name"`
			Path      int                    `json:"path"`
			Params    map[string]interface{} `json:"params"`
		} `json:"blocks"`
	}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(response.Blocks) != 1 {
		t.Fatalf("block count = %d, want 1", len(response.Blocks))
	}
	block := response.Blocks[0]
	if block.Name != "Drive" || block.ModelName != "Scream 808" || block.Path != 0 {
		t.Fatalf("block = %+v, want requested mapping", block)
	}
	if block.Params["Gain"] != 0.2 || block.Params["Tone"] != 0.4 {
		t.Fatalf("params = %v, want Gain 0.2 and Tone 0.4", block.Params)
	}
}
