package gemini

import (
	"HelAIx/pkg/helix"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// ChatPresetEngineer takes the abstract rig and maps it to specific Helix Blocks, or refines an existing implementation
func (c *Client) ChatPresetEngineer(ctx context.Context, rig *RigDescription, presetName string, history []ChatMessage, hardware string, defaultExp int, variaxEnabled bool, hardwareModel string) (*helix.Preset, error) {
	// 1. Prepare Catalog Context
	helix.DB.EnsureLoaded()

	// Create a list of available models with their mono DSP costs
	var availableModels strings.Builder
	for _, e := range helix.DB.Entries {
		cost := e.DSPMono
		if cost == 0 {
			cost = 3.0 // Reasonable default
		}
		availableModels.WriteString(fmt.Sprintf("- %s (Based on: %s) [DSP: %.1f%%]\n", e.Name, e.BasedOn, cost))
	}

	// 2. Hardware Capabilities
	isDualDSP := hardware == "Helix Floor" || hardware == "Helix LT" || hardware == "Helix Rack"
	dspCapacity := "1 path of 100%"
	if isDualDSP {
		dspCapacity = "2 paths (Path 1 and Path 2), each with its own 100% DSP chip. Total 200%."
	}

	// 3. System Prompt
	sysPrompt := fmt.Sprintf(`You are a Line 6 Helix expert and Preset Engineer. 
	You will receive a "Rig Description" (abstract design) and the conversation history.
	Your job is to MAP each item to the BEST MATCHING Available Helix Model from the provided list.
	
	TARGET HARDWARE: %s
	DSP CAPACITY: %s
	
	AVAILABLE MODELS:
	%s
	
	CONVERSATION LOGIC:
	- If the user provides feedback, adjust the technical implementation.
	- You MUST stick to the Gear list proposed by the Sound Engineer.
	
	DSP MANAGEMENT:
	- You MUST NOT exceed 100%% DSP per path.
	- Each path (Path 1 and Path 2) has its own 100%% budget.
	- Sum the "DSP" percentages of the blocks you place in each path.
	- If the hardware has 2 paths, you can distribute blocks between them to avoid clipping.
	- Path 1 is "path": 0, Path 2 is "path": 1.
	- If the hardware has only 1 path, use "path": 0 for everything.

	PARAMETER CONSTRAINTS:
	- For Reverb blocks, NEVER set "Decay" or "VerbDecay" to its maximum value (1.0). Keep it at 0.7 or lower to avoid excessive noise/feedback loops.

	OUTPUT INSTRUCTIONS:
	- Return ONLY a JSON object with a "blocks" array.
	- "model_name" must match a Name from the list.
	- "path" must be 0 (Path 1) or 1 (Path 2).
	
	OUTPUT FORMAT:
	{
		"blocks": [
			{ "id": "block0", "model_name": "US Deluxe Nrm", "path": 0, "params": { "Drive": 0.6 } },
			{ "id": "block1", "model_name": "Minotaur", "path": 0, "params": { "Gain": 0.2 } }
		]
	}
	`, hardware, dspCapacity, availableModels.String())

	// Truncate prompt if needed (though Gemini 1.5 Handle this well)
	if len(sysPrompt) > 100000 {
		sysPrompt = sysPrompt[:100000] + "... (truncated)"
	}

	// 4. Prepare User Input
	inputBytes, _ := json.Marshal(rig)
	userPrompt := string(inputBytes)

	// Construct the conversation history
	var contents []*genai.Content

	// Add system instruction and rig proposal
	contents = append(contents, &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: sysPrompt},
			{Text: fmt.Sprintf("SOUND ENGINEER PROPOSAL: %s", userPrompt)},
		},
	})

	// Add conversation history
	for _, msg := range history {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: msg.Content}},
		})
	}

	// Generate content with JSON response format
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := c.client.Models.GenerateContent(ctx, c.ModelName, contents, config)
	if err != nil {
		return nil, fmt.Errorf("preset engineer agent failed: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Preset Engineer Agent")
	}

	jsonText := resp.Candidates[0].Content.Parts[0].Text

	// 5. Parse Response
	type BuilderBlock struct {
		ModelName string                 `json:"model_name"`
		Path      int                    `json:"path"`
		Params    map[string]interface{} `json:"params"`
	}
	type BuilderResponse struct {
		Blocks []BuilderBlock `json:"blocks"`
	}

	var builderResp BuilderResponse
	if err := json.Unmarshal([]byte(jsonText), &builderResp); err != nil {
		return nil, fmt.Errorf("failed to parse Preset Engineer JSON: %v. Raw: %s", err, jsonText)
	}

	// 6. Construct The Real Preset via Template
	preset, err := helix.NewTemplatePreset(presetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create preset from template: %v", err)
	}

	// Helper to get dsp map
	getDSP := func(path int) map[string]interface{} {
		key := fmt.Sprintf("dsp%d", path)
		if data, ok := (*preset)["data"].(map[string]interface{}); ok {
			if tone, ok := data["tone"].(map[string]interface{}); ok {
				if d, ok := tone[key].(map[string]interface{}); ok {
					return d
				}
			}
		}
		return nil
	}

	// Helper to get controller map
	getController := func() map[string]interface{} {
		if data, ok := (*preset)["data"].(map[string]interface{}); ok {
			if tone, ok := data["tone"].(map[string]interface{}); ok {
				if ctrl, ok := tone["controller"].(map[string]interface{}); ok {
					if d, ok := ctrl["dsp0"].(map[string]interface{}); ok {
						return d
					}
				}
			}
		}
		return nil
	}

	// Loop and place blocks
	path0Count := 0
	path1Count := 0

	for _, b := range builderResp.Blocks {
		entry, found := helix.DB.FindByRealName(b.ModelName)
		var internalID string
		var defaultData map[string]interface{}

		if found {
			internalID = entry.InternalName
			defaultData = entry.Data["Defaults"].(map[string]interface{})
		} else {
			entryID, foundID := helix.DB.FindByID(b.ModelName)
			if foundID {
				internalID = entryID.InternalName
				defaultData = entryID.Data["Defaults"].(map[string]interface{})
			} else {
				continue
			}
		}

		finalParams := make(map[string]interface{})
		for k, v := range defaultData {
			finalParams[k] = v
		}
		finalParams["@model"] = internalID
		finalParams["@enabled"] = true

		// Target Path
		targetPath := b.Path
		if !isDualDSP {
			targetPath = 0
		}
		finalParams["@path"] = targetPath

		// Position
		pos := 0
		if targetPath == 1 {
			pos = path1Count
			path1Count++
		} else {
			pos = path0Count
			path0Count++
		}
		finalParams["@position"] = pos

		// Type mapping
		switch {
		case strings.HasPrefix(internalID, "HD2_Amp"):
			finalParams["@type"] = 1
		case strings.HasPrefix(internalID, "HD2_Preamp"):
			finalParams["@type"] = 2
		case strings.HasPrefix(internalID, "HD2_Cab") || strings.HasPrefix(internalID, "VIC_Cab"):
			finalParams["@type"] = 4
		case strings.HasPrefix(internalID, "HD2_Delay") || strings.HasPrefix(internalID, "HD2_Reverb") || strings.HasPrefix(internalID, "VIC_Reverb"):
			finalParams["@type"] = 7
		default:
			finalParams["@type"] = 0
		}

		for k, v := range b.Params {
			// Reverb Decay Sanitization (Reverbs and Delays with Reverb tails)
			isReverb := strings.HasPrefix(internalID, "HD2_Reverb") || strings.HasPrefix(internalID, "VIC_Reverb")
			isDelay := strings.HasPrefix(internalID, "HD2_Delay") || strings.HasPrefix(internalID, "VIC_Delay")
			if (isReverb && (k == "Decay" || k == "VerbDecay")) || (isDelay && k == "VerbDecay") {
				if val, ok := v.(float64); ok && val >= 0.7 {
					v = 0.7
				}
			}
			finalParams[k] = v
		}

		// Place in correct DSP
		dsp := getDSP(targetPath)
		if dsp != nil {
			blockKey := fmt.Sprintf("block%d", pos)
			dsp[blockKey] = finalParams

			// Default Expression Pedal Assignment
			if defaultExp > 0 {
				isWah := strings.HasPrefix(internalID, "HD2_Wah")
				isVol := strings.HasPrefix(internalID, "HD2_Vol")
				isPitch := strings.HasPrefix(internalID, "HD2_PitchPitchWham")
				if isWah || isVol || isPitch {
					ctrls := getController()
					if ctrls != nil {
						ctrls[blockKey] = map[string]interface{}{
							"Pedal": map[string]interface{}{
								"@controller":       defaultExp,
								"@max":              1.0,
								"@min":              0.0,
								"@snapshot_disable": false,
							},
						}
					}
				}
			}

			// Sync snapshot
			if data, ok := (*preset)["data"].(map[string]interface{}); ok {
				if tone, ok := data["tone"].(map[string]interface{}); ok {
					for s := 0; s < 8; s++ {
						snapKey := fmt.Sprintf("snapshot%d", s)
						if snap, ok := tone[snapKey].(map[string]interface{}); ok {
							if blocks, ok := snap["blocks"].(map[string]interface{}); ok {
								dspKey := fmt.Sprintf("dsp%d", targetPath)
								if targetDSP, ok := blocks[dspKey].(map[string]interface{}); ok {
									targetDSP[blockKey] = true
								}
							}
						}
					}
				}
			}
		}
	}

	// 5. Apply Variax Settings
	if variaxEnabled {
		applyVariax(preset, rig, hardwareModel)
	} else {
		// If Variax is disabled, reset to safe defaults instead of removing
		if data, ok := (*preset)["data"].(map[string]interface{}); ok {
			if tone, ok := data["tone"].(map[string]interface{}); ok {
				tone["variax"] = map[string]interface{}{
					"@model":               "@variax",
					"@variax_customtuning": false,
					"@variax_lockctrls":    0,
					"@variax_magmode":      true,
					"@variax_model":        0,
					"@variax_str1level":    1.0,
					"@variax_str1tuning":   0,
					"@variax_str2level":    1.0,
					"@variax_str2tuning":   0,
					"@variax_str3level":    1.0,
					"@variax_str3tuning":   0,
					"@variax_str4level":    1.0,
					"@variax_str4tuning":   0,
					"@variax_str5level":    1.0,
					"@variax_str5tuning":   0,
					"@variax_str6level":    1.0,
					"@variax_str6tuning":   0,
					"@variax_toneknob":     -0.10,
					"@variax_volumeknob":   -0.10,
				}
			}
		}
	}

	// 6. Apply Global Defaults (Cursor)
	if data, ok := (*preset)["data"].(map[string]interface{}); ok {
		if tone, ok := data["tone"].(map[string]interface{}); ok {
			if global, ok := tone["global"].(map[string]interface{}); ok {
				global["@cursor_dsp"] = 0
				global["@cursor_group"] = "inputA"
			}
		}
	}

	return preset, nil
}

func applyVariax(preset *helix.Preset, rig *RigDescription, hardwareModel string) {
	data, ok := (*preset)["data"].(map[string]interface{})
	if !ok {
		return
	}
	tone, ok := data["tone"].(map[string]interface{})
	if !ok {
		return
	}
	v, ok := tone["variax"].(map[string]interface{})
	if !ok {
		return
	}

	// 1. Model Selection (Real Name -> Variax Name)
	// User hardware maps T-Model to 10 (Custom banks at 0-9?), so we apply a +10 offset to standard Helix IDs.
	modelID := -1
	m := strings.ToLower(rig.GuitarModel)

	switch {
	// Telecaster / T-Model (Standard ID 0 -> Hardware ID 10)
	case strings.Contains(m, "tele") || strings.Contains(m, "t-model"):
		modelID = 10
	// Stratocaster / Spank (Standard ID 5 -> Hardware ID 15)
	case strings.Contains(m, "strat") || strings.Contains(m, "spank"):
		modelID = 15
	// Les Paul / Lester (Standard ID 10 -> Hardware ID 20)
	case strings.Contains(m, "paul") || strings.Contains(m, "lester") || strings.Contains(m, "lp") || strings.Contains(m, "gibson") || strings.Contains(m, "singlecut") || strings.Contains(m, "humbucker") || strings.Contains(m, "arm the homeless"):
		modelID = 20
	// Special / Junior / P90 (Standard ID 15 -> Hardware ID 25)
	case strings.Contains(m, "special") || strings.Contains(m, "junior") || strings.Contains(m, "p90"):
		modelID = 25
	// R-Billy / Gretsch (Standard ID 20 -> Hardware ID 30)
	case strings.Contains(m, "gretsch") || strings.Contains(m, "billy") || strings.Contains(m, "silver jet") || strings.Contains(m, "duo jet"):
		modelID = 30
	// Chime / Rickenbacker (Standard ID 25 -> Hardware ID 35)
	case strings.Contains(m, "rick") || strings.Contains(m, "chime"):
		modelID = 35
	// Semi / ES-335 (Standard ID 30 -> Hardware ID 40)
	case strings.Contains(m, "semi") || strings.Contains(m, "335") || strings.Contains(m, "es-335"):
		modelID = 40
	// Jazz / L5 (Standard ID 35 -> Hardware ID 45)
	case strings.Contains(m, "jazz") || strings.Contains(m, "175") || strings.Contains(m, "super 400") || strings.Contains(m, "l5"):
		modelID = 45
	// Acoustic (Standard ID 40 -> Hardware ID 50)
	case strings.Contains(m, "acoustic") || strings.Contains(m, "j-200") || strings.Contains(m, "d-28") || strings.Contains(m, "martin") || strings.Contains(m, "taylor"):
		modelID = 50
	// Reso / Dobro (Standard ID 45 -> Hardware ID 55)
	case strings.Contains(m, "reso") || strings.Contains(m, "dobro") || strings.Contains(m, "banjo") || strings.Contains(m, "sitar"):
		modelID = 55
	// Shuriken (Standard ID 50 -> Hardware ID 60? Or keep distinct?)
	// Assuming +10 offset applies linearly relative to standard banks.
	case strings.Contains(m, "shuriken") && strings.ToLower(hardwareModel) == "shuriken":
		modelID = 60
	}

	// If a model was found, apply it
	if modelID >= 0 {
		v["@variax_model"] = modelID
		v["@variax_magmode"] = true
		// Store the technical name in metadata for the visualizer to use later
		// (We'll use a custom property in the preset name or similar if Helix supports it,
		// or just rely on the mapping for the frontend)
	} else {
		// If no specific model suggested, don't force Variax settings
		v["@variax_magmode"] = true
	}

	// 2. Tuning Logic
	t := strings.ToLower(rig.Tuning)
	if t != "" && t != "standard" {
		offsets := [6]int{0, 0, 0, 0, 0, 0}
		isTuningMapped := false

		switch {
		case strings.Contains(t, "drop d"):
			offsets = [6]int{-2, 0, 0, 0, 0, 0} // Low E to D
			isTuningMapped = true
		case strings.Contains(t, "eb") || strings.Contains(t, "half step down"):
			offsets = [6]int{-1, -1, -1, -1, -1, -1}
			isTuningMapped = true
		case strings.Contains(t, "d standard") || strings.Contains(t, "whole step down"):
			offsets = [6]int{-2, -2, -2, -2, -2, -2}
			isTuningMapped = true
		case strings.Contains(t, "drop c") && (strings.ToLower(hardwareModel) == "shuriken" || strings.ToLower(hardwareModel) == "jtv"):
			offsets = [6]int{-4, -2, -2, -2, -2, -2}
			isTuningMapped = true
		case strings.Contains(t, "drop b"):
			// Slipknot-style Drop B: B-F#-B-E-G#-C#
			offsets = [6]int{-5, -3, -3, -3, -3, -3}
			isTuningMapped = true
		case strings.Contains(t, "drop a"):
			offsets = [6]int{-7, -5, -5, -5, -5, -5}
			isTuningMapped = true
		case strings.Contains(t, "baritone"):
			offsets = [6]int{-5, -5, -5, -5, -5, -5} // B Standard
			isTuningMapped = true
		case strings.Contains(t, "open g"):
			offsets = [6]int{-2, -2, 0, 0, 0, -2} // D-G-D-G-B-D
			isTuningMapped = true
		case strings.Contains(t, "open d"):
			offsets = [6]int{-2, 0, 0, -1, -2, -2} // D-A-D-F#-A-D
			isTuningMapped = true
		case strings.Contains(t, "dadgad"):
			offsets = [6]int{-2, 0, 0, 0, 0, -2}
			isTuningMapped = true
		}

		if isTuningMapped {
			v["@variax_customtuning"] = true
			v["@variax_str1tuning"] = offsets[5]
			v["@variax_str2tuning"] = offsets[4]
			v["@variax_str3tuning"] = offsets[3]
			v["@variax_str4tuning"] = offsets[2]
			v["@variax_str5tuning"] = offsets[1]
			v["@variax_str6tuning"] = offsets[0]
		}
	} else {
		v["@variax_customtuning"] = false
	}
}
