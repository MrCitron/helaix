package gemini

import (
	"HelAIx/pkg/helix"
	"fmt"
	"strings"
	"testing"
)

func TestTemplatePresetParametersSatisfyContracts(t *testing.T) {
	preset, err := helix.NewTemplatePreset("Contract Compatibility")
	if err != nil {
		t.Fatalf("NewTemplatePreset() error = %v", err)
	}

	data := (*preset)["data"].(map[string]interface{})
	tone := data["tone"].(map[string]interface{})
	for _, dspName := range []string{"dsp0", "dsp1"} {
		dsp := tone[dspName].(map[string]interface{})
		for blockName, rawBlock := range dsp {
			if !strings.HasPrefix(blockName, "block") {
				continue
			}
			block := rawBlock.(map[string]interface{})
			modelID, ok := block["@model"].(string)
			if !ok {
				t.Fatalf("%s.%s has no model", dspName, blockName)
			}
			entry, ok := helix.DB.FindByID(modelID)
			if !ok {
				t.Fatalf("%s.%s uses unknown model %q", dspName, blockName, modelID)
			}
			params := make(map[string]interface{})
			for name, value := range block {
				if !strings.HasPrefix(name, "@") {
					params[name] = value
				}
			}
			if err := validateParameterMap(entry, params, false); err != nil {
				t.Error(fmt.Errorf("%s.%s model %q: %w", dspName, blockName, modelID, err))
			}
		}
	}
}

func TestCatalogDefaultsSatisfyBaselineContracts(t *testing.T) {
	helix.DB.EnsureLoaded()
	for _, entry := range helix.DB.Entries {
		defaults, ok := entry.Data["Defaults"].(map[string]interface{})
		if !ok {
			t.Fatalf("model %q has invalid defaults", entry.InternalName)
		}
		params := make(map[string]interface{})
		for name, value := range defaults {
			if !strings.HasPrefix(name, "@") {
				params[name] = value
			}
		}
		if err := validateCatalogDefaultParameterMap(entry, params); err != nil {
			t.Error(fmt.Errorf("model %q defaults: %w", entry.InternalName, err))
		}
	}
}
