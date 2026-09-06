package codexcli

import (
	"HelAIx/pkg/gemini"
	"context"
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
