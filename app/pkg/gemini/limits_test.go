package gemini

import (
	"HelAIx/pkg/helix"
	"math"
	"strings"
	"testing"
)

func TestParameterContractsResolveAliasesAndValidateTypes(t *testing.T) {
	entry := contractTestEntry("HD2_AmpTest")

	name, err := resolveParameterName(entry, "gain", false)
	if err != nil || name != "Drive" {
		t.Fatalf("resolveParameterName(gain) = %q, %v; want Drive, nil", name, err)
	}

	tests := []struct {
		name   string
		values map[string]interface{}
		want   string
	}{
		{name: "valid", values: map[string]interface{}{"gain": 0.5}},
		{name: "reserved", values: map[string]interface{}{"@model": "other"}, want: "reserved"},
		{name: "wrong type", values: map[string]interface{}{"Drive": "0.5"}, want: "must be a number"},
		{name: "non finite", values: map[string]interface{}{"Drive": math.NaN()}, want: "must be finite"},
		{name: "outside controller range", values: map[string]interface{}{"Drive": 1.1}, want: "outside the allowed range"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParameterMap(entry, tt.values, false)
			if tt.want == "" && err != nil {
				t.Fatalf("validateParameterMap() error = %v", err)
			}
			if tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)) {
				t.Fatalf("validateParameterMap() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestParameterContractsApplySnapshotAndSafetyLimits(t *testing.T) {
	entry := contractTestEntry("HD2_ReverbTest")
	entry.Data["Defaults"].(map[string]interface{})["Decay"] = 0.5
	entry.Data["Controller_Dict"].(map[string]interface{})["Decay"] = map[string]interface{}{"@min": 0.0, "@max": 1.0}

	if err := validateParameterMap(entry, map[string]interface{}{"Decay": 0.8}, true); err == nil || !strings.Contains(err.Error(), "outside the allowed range") {
		t.Fatalf("reverb safety error = %v, want range rejection", err)
	}
	if err := validateParameterMap(entry, map[string]interface{}{"Mode": "Modern"}, true); err == nil || !strings.Contains(err.Error(), "cannot be controlled by snapshots") {
		t.Fatalf("snapshot control error = %v, want controllability rejection", err)
	}

	delay := contractTestEntry("HD2_DelayTest")
	delay.Data["Defaults"].(map[string]interface{})["Feedback"] = 0.5
	delay.Data["Controller_Dict"].(map[string]interface{})["Feedback"] = map[string]interface{}{"@min": 0.0, "@max": 1.0}
	if err := validateParameterMap(delay, map[string]interface{}{"Feedback": 0.8}, false); err == nil || !strings.Contains(err.Error(), "outside the allowed range") {
		t.Fatalf("delay safety error = %v, want range rejection", err)
	}
}

func contractTestEntry(internalName string) helix.CatalogEntry {
	return helix.CatalogEntry{
		InternalName: internalName,
		Data: map[string]interface{}{
			"Defaults": map[string]interface{}{
				"@model": internalName,
				"Drive":  0.5,
				"Mode":   "Modern",
			},
			"Controller_Dict": map[string]interface{}{
				"Drive": map[string]interface{}{"@min": 0.0, "@max": 1.0},
			},
		},
	}
}
