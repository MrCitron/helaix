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
func (c *Client) ChatPresetEngineer(ctx context.Context, rig *RigDescription, presetName string, history []ChatMessage, hardware string, defaultExp int) (*helix.Preset, error) {
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

	return preset, nil
}
