package gemini

import (
	"HelAIx/pkg/helix"
	"encoding/json"
	"strings"
	"testing"
)

func TestVariaxConfigurationIsEmbedded(t *testing.T) {
	if len(helix.VariaxModelsJSON) == 0 {
		t.Fatal("VariaxModelsJSON is empty")
	}

	var config map[string]interface{}
	if err := json.Unmarshal(helix.VariaxModelsJSON, &config); err != nil {
		t.Fatalf("embedded Variax configuration is invalid: %v", err)
	}
}

func TestValidateBuilderResponseRejectsUnknownModel(t *testing.T) {
	rig := &RigDescription{
		Chain: []RigComponent{{Name: "Amp", Type: "amp"}},
	}
	response := builderResponse{
		Blocks: []builderBlock{{Name: "Amp", ModelName: "Not A Helix Model", Path: 0}},
	}

	err := validateBuilderResponse(response, rig, true)
	if err == nil || !strings.Contains(err.Error(), "unknown model") {
		t.Fatalf("validateBuilderResponse() error = %v, want unknown model error", err)
	}
}

func TestValidateBuilderResponseRequiresEveryRigComponent(t *testing.T) {
	rig := &RigDescription{
		Chain: []RigComponent{
			{Name: "Amp", Type: "amp"},
			{Name: "Cab", Type: "cab"},
		},
	}
	response := builderResponse{
		Blocks: []builderBlock{{Name: "Amp", ModelName: "A30 Fawn Brt", Path: 0}},
	}

	err := validateBuilderResponse(response, rig, true)
	if err == nil || !strings.Contains(err.Error(), "expected 2 blocks") {
		t.Fatalf("validateBuilderResponse() error = %v, want missing block error", err)
	}
}

func TestValidatePresetRejectsUnknownModel(t *testing.T) {
	preset, err := helix.NewTemplatePreset("Test")
	if err != nil {
		t.Fatalf("NewTemplatePreset() error = %v", err)
	}

	data := (*preset)["data"].(map[string]interface{})
	tone := data["tone"].(map[string]interface{})
	dsp0 := tone["dsp0"].(map[string]interface{})
	dsp0["block0"] = map[string]interface{}{"@model": "unknown"}

	err = helix.ValidatePreset(*preset)
	if err == nil || !strings.Contains(err.Error(), "invalid model") {
		t.Fatalf("ValidatePreset() error = %v, want invalid model error", err)
	}
}

func TestValidatePresetAcceptsGeminiPresetWithCoherentSnapshots(t *testing.T) {
	preset, err := helix.NewTemplatePreset("Gemini Snapshot Test")
	if err != nil {
		t.Fatalf("NewTemplatePreset() error = %v", err)
	}

	data := (*preset)["data"].(map[string]interface{})
	tone := data["tone"].(map[string]interface{})
	tone["dsp0"].(map[string]interface{})["block0"] = map[string]interface{}{
		"@model": "HD2_AmpA30FawnBrt",
	}
	tone["dsp1"].(map[string]interface{})["block0"] = map[string]interface{}{
		"@model": "HD2_CabMicIr_1x10USPrincess",
	}

	for _, snapshotKey := range []string{"snapshot0", "snapshot1"} {
		snapshot := tone[snapshotKey].(map[string]interface{})
		blocks := snapshot["blocks"].(map[string]interface{})
		blocks["dsp0"] = map[string]interface{}{"block0": true}
		blocks["dsp1"] = map[string]interface{}{"block0": false}
	}

	if err := helix.ValidatePreset(*preset); err != nil {
		t.Fatalf("ValidatePreset() error = %v, want nil", err)
	}
}

func TestValidatePresetRejectsSnapshotReferenceToMissingBlock(t *testing.T) {
	preset, err := helix.NewTemplatePreset("Invalid Snapshot Test")
	if err != nil {
		t.Fatalf("NewTemplatePreset() error = %v", err)
	}

	data := (*preset)["data"].(map[string]interface{})
	tone := data["tone"].(map[string]interface{})
	tone["dsp0"].(map[string]interface{})["block0"] = map[string]interface{}{
		"@model": "HD2_AmpA30FawnBrt",
	}
	snapshot := tone["snapshot0"].(map[string]interface{})
	snapshot["blocks"] = map[string]interface{}{
		"dsp0": map[string]interface{}{"block1": true},
	}

	err = helix.ValidatePreset(*preset)
	if err == nil || !strings.Contains(err.Error(), "references missing block") {
		t.Fatalf("ValidatePreset() error = %v, want missing snapshot block error", err)
	}
}
