package helix

import (
	"fmt"
	"strings"
)

// ValidatePreset verifies the generated preset has valid Helix effect blocks
// before it is returned to the UI or written to disk.
func ValidatePreset(preset Preset) error {
	data, ok := preset["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("preset is missing data")
	}
	tone, ok := data["tone"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("preset is missing tone data")
	}

	blockCount := 0
	blockKeys := map[string]map[string]struct{}{
		"dsp0": {},
		"dsp1": {},
	}
	for _, dspKey := range []string{"dsp0", "dsp1"} {
		dsp, ok := tone[dspKey].(map[string]interface{})
		if !ok {
			return fmt.Errorf("preset is missing %s", dspKey)
		}

		for key, rawBlock := range dsp {
			if !strings.HasPrefix(key, "block") {
				continue
			}

			block, ok := rawBlock.(map[string]interface{})
			if !ok {
				return fmt.Errorf("%s.%s is not a block object", dspKey, key)
			}
			modelID, ok := block["@model"].(string)
			if !ok || !IsValidModel(modelID) {
				return fmt.Errorf("%s.%s has an invalid model", dspKey, key)
			}
			blockKeys[dspKey][key] = struct{}{}
			blockCount++
		}
	}

	if blockCount == 0 {
		return fmt.Errorf("preset has no effect blocks")
	}
	if err := validateSnapshots(tone, blockKeys); err != nil {
		return err
	}
	return nil
}

func validateSnapshots(tone map[string]interface{}, blockKeys map[string]map[string]struct{}) error {
	for index := 0; index < 8; index++ {
		snapshotKey := fmt.Sprintf("snapshot%d", index)
		rawSnapshot, exists := tone[snapshotKey]
		if !exists {
			continue
		}
		snapshot, ok := rawSnapshot.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s is not an object", snapshotKey)
		}
		rawBlocks, exists := snapshot["blocks"]
		if !exists {
			continue
		}
		blocks, ok := rawBlocks.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s.blocks is not an object", snapshotKey)
		}

		for dspKey, rawStates := range blocks {
			knownBlocks, knownDSP := blockKeys[dspKey]
			if !knownDSP {
				return fmt.Errorf("%s references unknown path %s", snapshotKey, dspKey)
			}
			states, ok := rawStates.(map[string]interface{})
			if !ok {
				return fmt.Errorf("%s.blocks.%s is not an object", snapshotKey, dspKey)
			}
			for blockKey, state := range states {
				if _, exists := knownBlocks[blockKey]; !exists {
					return fmt.Errorf("%s references missing block %s.%s", snapshotKey, dspKey, blockKey)
				}
				if _, ok := state.(bool); !ok {
					return fmt.Errorf("%s has non-boolean state for %s.%s", snapshotKey, dspKey, blockKey)
				}
			}
		}
	}

	return nil
}
