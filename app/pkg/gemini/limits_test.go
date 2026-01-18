package gemini

import (
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
