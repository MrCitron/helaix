package helix

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed data/template.json
var templateJSON []byte

// Preset is now a generic map because the Native schema is too complex/undocumented
// to strictly type without risk of missing fields.
type Preset map[string]interface{}

// NewTemplatePreset returns a fresh copy of the template preset
// with the Name updated and Effects blocks/controllers/footswitches cleared.
func NewTemplatePreset(name string) (*Preset, error) {
	var p Preset
	if err := json.Unmarshal(templateJSON, &p); err != nil {
		return nil, fmt.Errorf("failed to load template: %v", err)
	}

	// 1. Update Name
	if data, ok := p["data"].(map[string]interface{}); ok {
		if meta, ok := data["meta"].(map[string]interface{}); ok {
			meta["name"] = name
		}
	}

	// 2. Clear Blocks, Controllers, and Footswitches in Tone
	if data, ok := p["data"].(map[string]interface{}); ok {
		if tone, ok := data["tone"].(map[string]interface{}); ok {
			// Clear dsp0 effect blocks
			if dsp0, ok := tone["dsp0"].(map[string]interface{}); ok {
				for k := range dsp0 {
					if strings.HasPrefix(k, "block") || strings.HasPrefix(k, "cab") {
						delete(dsp0, k)
					}
				}
			}

			// Clear Controllers
			if ctrl, ok := tone["controller"].(map[string]interface{}); ok {
				if dsp0, ok := ctrl["dsp0"].(map[string]interface{}); ok {
					for k := range dsp0 {
						delete(dsp0, k)
					}
				}
			}

			// Clear Footswitches
			if fs, ok := tone["footswitch"].(map[string]interface{}); ok {
				if dsp0, ok := fs["dsp0"].(map[string]interface{}); ok {
					for k := range dsp0 {
						delete(dsp0, k)
					}
				}
			}

			// 3. Clear Snapshots (Surgical Reset)
			// Ensure snapshots don't have dangling references to deleted blocks.
			for i := 0; i < 8; i++ {
				snapKey := fmt.Sprintf("snapshot%d", i)
				if snap, ok := tone[snapKey].(map[string]interface{}); ok {
					// Clear blocks map in snapshot
					if blocks, ok := snap["blocks"].(map[string]interface{}); ok {
						if dsp0, ok := blocks["dsp0"].(map[string]interface{}); ok {
							for k := range dsp0 {
								delete(dsp0, k)
							}
						}
					}
					// Clear controllers map in snapshot
					if ctrls, ok := snap["controllers"].(map[string]interface{}); ok {
						if dsp0, ok := ctrls["dsp0"].(map[string]interface{}); ok {
							for k := range dsp0 {
								delete(dsp0, k)
							}
						}
					}
				}
			}
		}
	}

	return &p, nil
}
