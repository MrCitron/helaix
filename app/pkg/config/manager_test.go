package config

import (
	"HelAIx/pkg/provider"
	"testing"
)

func TestNormalizeProviderMigratesGoogleSettings(t *testing.T) {
	if got := NormalizeProvider("Google"); got != provider.Gemini {
		t.Fatalf("NormalizeProvider(Google) = %q, want %q", got, provider.Gemini)
	}
	if got := NormalizeProvider(""); got != provider.Gemini {
		t.Fatalf("NormalizeProvider(empty) = %q, want %q", got, provider.Gemini)
	}
}

func TestIsSupportedProvider(t *testing.T) {
	for _, name := range []string{provider.Gemini, provider.CodexCLI} {
		if !IsSupportedProvider(name) {
			t.Fatalf("IsSupportedProvider(%q) = false", name)
		}
	}
	if got := NormalizeProvider("codex_oauth"); got != provider.CodexCLI {
		t.Fatalf("NormalizeProvider(codex_oauth) = %q, want %q", got, provider.CodexCLI)
	}
	if IsSupportedProvider("unknown") {
		t.Fatal("IsSupportedProvider accepted an unknown provider")
	}
}
