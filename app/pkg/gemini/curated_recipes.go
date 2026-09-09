package gemini

import (
	"encoding/json"
	"strings"
)

type curatedRecipeBlock struct {
	componentType string
	modelName     string
	params        map[string]interface{}
}

type curatedToneRecipe struct {
	name    string
	signals []string
	blocks  []curatedRecipeBlock
}

var curatedToneRecipes = []curatedToneRecipe{
	{
		name:    "clean",
		signals: []string{"clean", "limpio", "cristalino"},
		blocks: []curatedRecipeBlock{
			{componentType: "amp", modelName: "US Deluxe Nrm", params: map[string]interface{}{"Drive": 0.25, "Bass": 0.50, "Mid": 0.55, "Treble": 0.60}},
			{componentType: "cab", modelName: "1x12 US Deluxe", params: map[string]interface{}{"Distance": 3.0, "LowCut": 90.0, "HighCut": 9000.0, "Mic": 1.0, "Position": 0.35}},
			{componentType: "reverb", modelName: "HD2_Reverb63Spring", params: map[string]interface{}{"Decay": 0.32, "Mix": 0.18, "HighCut": 6500.0}},
		},
	},
	{
		name:    "crunch",
		signals: []string{"crunch", "crujiente", "brit crunch"},
		blocks: []curatedRecipeBlock{
			{componentType: "pedal", modelName: "Prize Drive", params: map[string]interface{}{"Bass Cut": true, "Drive": 0.18, "Level": 0.65, "Spectrum": 0.50}},
			{componentType: "amp", modelName: "Brit Plexi Jump", params: map[string]interface{}{"BrtDrive": 0.48, "NrmDrive": 0.42, "Bass": 0.42, "Mid": 0.62, "Treble": 0.55}},
			{componentType: "cab", modelName: "4x12 Brit V30", params: map[string]interface{}{"Distance": 2.0, "LowCut": 90.0, "HighCut": 8500.0, "Mic": 2.0}},
			{componentType: "reverb", modelName: "HD2_ReverbRoom", params: map[string]interface{}{"Decay": 0.22, "Mix": 0.12, "HighCut": 5000.0}},
		},
	},
	{
		name:    "tight-metal",
		signals: []string{"tight metal", "metal", "djent", "heavy"},
		blocks: []curatedRecipeBlock{
			{componentType: "pedal", modelName: "Scream 808", params: map[string]interface{}{"Gain": 0.10, "Tone": 0.58}},
			{componentType: "amp", modelName: "Cali Rectifire", params: map[string]interface{}{"Drive": 0.55, "Bass": 0.35, "Mid": 0.48, "Treble": 0.58, "Presence": 0.42}},
			{componentType: "cab", modelName: "4x12 Cali V30", params: map[string]interface{}{"Distance": 1.0, "LowCut": 95.0, "HighCut": 8000.0, "Mic": 10.0, "Position": 0.20}},
			{componentType: "reverb", modelName: "HD2_ReverbRoom", params: map[string]interface{}{"Decay": 0.16, "Mix": 0.08, "HighCut": 4200.0}},
		},
	},
	{
		name:    "ambient",
		signals: []string{"ambient", "ambiental", "ethereal", "etereo", "etéreo", "shimmer"},
		blocks: []curatedRecipeBlock{
			{componentType: "amp", modelName: "US Deluxe Nrm", params: map[string]interface{}{"Drive": 0.20, "Bass": 0.48, "Mid": 0.52, "Treble": 0.62}},
			{componentType: "cab", modelName: "1x12 US Deluxe", params: map[string]interface{}{"Distance": 3.0, "LowCut": 100.0, "HighCut": 9500.0, "Mic": 1.0, "Position": 0.40}},
			{componentType: "modulation", modelName: "70s Chorus", params: map[string]interface{}{"ChorusIntensity": 0.32, "Mix": 0.28, "VibratoRate": 0.25}},
			{componentType: "delay", modelName: "Cosmos Echo", params: map[string]interface{}{"Time": 0.62, "Feedback": 0.48, "Mix": 0.32, "WowFlutter": 0.25}},
			{componentType: "reverb", modelName: "Glitz", params: map[string]interface{}{"Decay": 0.68, "Mix": 0.38, "LowCut": 140.0, "HighCut": 7500.0, "Mod Mix": 0.42}},
		},
	},
	{
		name:    "acoustic",
		signals: []string{"acoustic", "acustic", "acústic", "piezo"},
		blocks: []curatedRecipeBlock{
			{componentType: "pedal", modelName: "Parametric", params: map[string]interface{}{"LowCut": 80.0, "LowFreq": 180.0, "LowGain": -2.0, "MidFreq": 2800.0, "MidGain": -2.5, "HighCut": 12000.0}},
			{componentType: "amp", modelName: "US Deluxe Nrm", params: map[string]interface{}{"Drive": 0.15, "Bass": 0.46, "Mid": 0.50, "Treble": 0.58}},
			{componentType: "cab", modelName: "1x12 US Deluxe", params: map[string]interface{}{"Distance": 4.0, "LowCut": 85.0, "HighCut": 11000.0, "Mic": 1.0, "Position": 0.45}},
			{componentType: "reverb", modelName: "HD2_ReverbRoom", params: map[string]interface{}{"Decay": 0.28, "Mix": 0.16, "HighCut": 7000.0}},
		},
	},
}

func curatedRecipeForRig(rig RigDescription) *curatedToneRecipe {
	parts := []string{rig.SuggestedName, rig.Explanation}
	for _, component := range rig.Chain {
		parts = append(parts, component.Name, component.Description, component.Settings)
	}
	text := strings.ToLower(strings.Join(parts, " "))
	for _, name := range []string{"tight-metal", "ambient", "acoustic", "crunch", "clean"} {
		for i := range curatedToneRecipes {
			if curatedToneRecipes[i].name != name {
				continue
			}
			for _, signal := range curatedToneRecipes[i].signals {
				if strings.Contains(text, signal) {
					return &curatedToneRecipes[i]
				}
			}
		}
	}
	return nil
}

func (r *curatedToneRecipe) blockFor(componentType string) (curatedRecipeBlock, bool) {
	if r == nil {
		return curatedRecipeBlock{}, false
	}
	for _, block := range r.blocks {
		if block.componentType == componentType {
			return block, true
		}
	}
	return curatedRecipeBlock{}, false
}

func cloneParams(params map[string]interface{}) map[string]interface{} {
	if len(params) == 0 {
		return nil
	}
	cloned := make(map[string]interface{}, len(params))
	for key, value := range params {
		cloned[key] = value
	}
	return cloned
}

func applyCuratedRecipeDefaults(response *builderResponse, rig *RigDescription) {
	recipe := curatedRecipeForRig(*rig)
	if recipe == nil {
		return
	}
	componentTypes := make(map[string]string, len(rig.Chain))
	for _, component := range rig.Chain {
		componentTypes[component.Name] = component.Type
	}
	for i := range response.Blocks {
		block := &response.Blocks[i]
		preferred, ok := recipe.blockFor(componentTypes[block.Name])
		if !ok || !strings.EqualFold(block.ModelName, preferred.modelName) {
			continue
		}
		params := cloneParams(preferred.params)
		for name, value := range block.Params {
			params[name] = value
		}
		block.Params = params
	}
}

func plannedParamsText(params map[string]interface{}) string {
	if len(params) == 0 {
		return "{}"
	}
	encoded, err := json.Marshal(params)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}
