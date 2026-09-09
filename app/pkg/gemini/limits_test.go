package gemini

import (
	"HelAIx/pkg/helix"
	"testing"
)

func TestSanitizeParam(t *testing.T) {
	tests := []struct {
		name       string
		internalID string
		key        string
		val        interface{}
		want       interface{}
	}{
		// Delay Feedback Tests
		{"Delay Feedback Normal", "HD2_DelayCompulsive", "Feedback", 0.5, 0.5},
		{"Delay Feedback High", "HD2_DelayCompulsive", "Feedback", 0.9, 0.75},
		{"Delay Fdbk High", "HD2_DelaySimple", "Fdbk", 0.8, 0.75},
		{"Legacy Delay Bk High", "HD2_DelayCosmos", "Bk", 1.0, 0.75},
		{"Not Delay Feedback High", "HD2_AmpUSDeluxe", "Feedback", 0.9, 0.9},

		// Reverb Decay Tests (Regression Check)
		{"Reverb Decay Normal", "HD2_ReverbHall", "Decay", 0.5, 0.5},
		{"Reverb Decay High", "HD2_ReverbHall", "Decay", 0.8, 0.7},
		{"Delay with VerbDecay High", "HD2_DelayTransistor", "VerbDecay", 0.9, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeParam(tt.internalID, tt.key, tt.val); got != tt.want {
				t.Errorf("sanitizeParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveParamKey(t *testing.T) {
	defaults := map[string]interface{}{
		"31p25Hz": 0.0,
		"62p5Hz":  0.0,
		"Drive":   0.0,
	}

	tests := []struct {
		requested string
		want      string
		ok        bool
	}{
		{"31Hz", "31p25Hz", true},
		{"62Hz", "62p5Hz", true},
		{"drive", "Drive", true},
		{"not-a-parameter", "", false},
	}
	for _, tt := range tests {
		got, ok := resolveParamKey(defaults, tt.requested)
		if got != tt.want || ok != tt.ok {
			t.Errorf("resolveParamKey(%q) = %q, %v; want %q, %v", tt.requested, got, ok, tt.want, tt.ok)
		}
	}
}

func TestNormalizeParamType(t *testing.T) {
	entry := helix.CatalogEntry{
		Data: map[string]interface{}{
			"Defaults": map[string]interface{}{
				"Mode": true,
			},
		},
	}

	if got := normalizeParamType(entry, "Mode", "BP"); got != true {
		t.Errorf("normalizeParamType() = %v, want true", got)
	}
	if got := normalizeParamType(entry, "Mode", false); got != false {
		t.Errorf("normalizeParamType() = %v, want false", got)
	}
}

func TestSnapshotControllerUsesHardwareID(t *testing.T) {
	controller := snapshotController(helix.CatalogEntry{}, "Drive")
	if controller["@controller"] != 19 {
		t.Errorf("snapshot controller ID = %v, want 19", controller["@controller"])
	}
}
