package gemini

import (
	"HelAIx/pkg/helix"
	"encoding/json"
	"fmt"
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

func TestValidateBuilderResponseRejectsIncompatibleModel(t *testing.T) {
	delay := helix.CandidatesForComponent("delay")[0]
	rig := &RigDescription{
		Chain: []RigComponent{{Name: "Amp", Type: "amp"}},
	}
	response := builderResponse{
		Blocks: []builderBlock{{Name: "Amp", ModelName: delay.Name, Path: 0}},
	}
	if response.Blocks[0].ModelName == "" {
		response.Blocks[0].ModelName = delay.InternalName
	}

	err := validateBuilderResponse(response, rig, true)
	if err == nil || !strings.Contains(err.Error(), "incompatible") {
		t.Fatalf("validateBuilderResponse() error = %v, want incompatible model error", err)
	}
}

func TestValidateBuilderResponseRejectsOverDSPBudget(t *testing.T) {
	pedal := helix.CandidatesForComponent("pedal")[0]
	modelName := pedal.Name
	if modelName == "" {
		modelName = pedal.InternalName
	}
	blockCount := int(helix.SafeDSPPerPath/helix.EffectiveDSPMono(pedal)) + 1
	rig := &RigDescription{Chain: make([]RigComponent, 0, blockCount)}
	response := builderResponse{Blocks: make([]builderBlock, 0, blockCount)}
	for i := 0; i < blockCount; i++ {
		name := fmt.Sprintf("Pedal %d", i)
		rig.Chain = append(rig.Chain, RigComponent{Name: name, Type: "pedal"})
		response.Blocks = append(response.Blocks, builderBlock{Name: name, ModelName: modelName, Path: 0})
	}

	err := validateBuilderResponse(response, rig, true)
	if err == nil || !strings.Contains(err.Error(), "exceeding the safe") {
		t.Fatalf("validateBuilderResponse() error = %v, want DSP budget error", err)
	}
}

func TestValidateRigDescriptionRejectsInvalidSignalChain(t *testing.T) {
	base := func() *RigDescription {
		return &RigDescription{
			SuggestedName: "TEST RIG",
			Explanation:   "A valid test signal chain.",
			GuitarModel:   "Fender Stratocaster",
			Tuning:        "Standard",
			Chain: []RigComponent{
				{Type: "pedal", Name: "Drive", Description: "Boost", Settings: "Low gain"},
				{Type: "amp", Name: "Amp", Description: "Core amp", Settings: "Clean"},
				{Type: "cab", Name: "Cab", Description: "Speaker", Settings: "2x12"},
				{Type: "delay", Name: "Delay", Description: "Echo", Settings: "Short"},
			},
			Snapshots: []Snapshot{{Name: "Main", ActiveBlocks: []string{"Drive", "Amp", "Cab", "Delay"}}},
		}
	}

	tests := []struct {
		name string
		edit func(*RigDescription)
		want string
	}{
		{
			name: "second amp",
			edit: func(rig *RigDescription) {
				rig.Chain = append(rig.Chain[:2], append([]RigComponent{{Type: "amp", Name: "Amp 2", Description: "Second", Settings: "Lead"}}, rig.Chain[2:]...)...)
			},
			want: "exactly one amp",
		},
		{
			name: "out of order post effect",
			edit: func(rig *RigDescription) {
				rig.Chain[0], rig.Chain[3] = rig.Chain[3], rig.Chain[0]
			},
			want: "out of order",
		},
		{
			name: "unused snapshot component",
			edit: func(rig *RigDescription) {
				rig.Snapshots[0].ActiveBlocks = []string{"Drive", "Amp", "Cab"}
			},
			want: "not active in any snapshot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rig := base()
			tt.edit(rig)
			err := validateRigDescription(rig)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validateRigDescription() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestCandidateCatalogForRigExcludesOtherBlockFamilies(t *testing.T) {
	catalog, err := CandidateCatalogForRig(RigDescription{
		Chain: []RigComponent{{Type: "amp", Name: "Amp"}},
	})
	if err != nil {
		t.Fatalf("CandidateCatalogForRig() error = %v", err)
	}
	if !strings.Contains(catalog, "Amp [amp, ranked by tone intent]:") {
		t.Fatalf("candidate catalog does not identify the amp component: %s", catalog)
	}
	if count := strings.Count(catalog, "\n- "); count != helix.PromptCandidateLimit {
		t.Fatalf("candidate catalog contains %d models, want %d", count, helix.PromptCandidateLimit)
	}
	if strings.Contains(catalog, "Scream 808") {
		t.Fatalf("amp candidates included pedal model Scream 808")
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
