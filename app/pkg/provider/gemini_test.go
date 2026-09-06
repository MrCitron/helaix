package provider

import "testing"

func TestNewGeminiRequiresAPIKey(t *testing.T) {
	if _, err := NewGemini("", "gemini-2.5-flash"); err == nil {
		t.Fatal("NewGemini accepted an empty API key")
	}
}
