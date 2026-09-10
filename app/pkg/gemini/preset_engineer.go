package gemini

import (
	"HelAIx/pkg/helix"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"google.golang.org/genai"
)

// ChatPresetEngineer takes the abstract rig and maps it to specific Helix Blocks, or refines an existing implementation
func (c *Client) ChatPresetEngineer(ctx context.Context, rig *RigDescription, presetName string, history []ChatMessage, hardware string, defaultExp int, variaxEnabled bool, hardwareModel string, defaultInstrument string) (*helix.Preset, error) {
	// 1. Prepare Catalog Context
	helix.DB.EnsureLoaded()

	// Create a list of available models with their mono DSP costs
	var availableModels strings.Builder
	for _, e := range helix.DB.Entries {
		cost := e.DSPMono
		if cost == 0 {
			cost = 3.0 // Reasonable default
		}
		instrument := e.Instrument
		if instrument == "" {
			instrument = "General"
		}
		availableModels.WriteString(fmt.Sprintf("- %s (Based on: %s) [Instrument: %s] [DSP: %.1f%%]\n", e.Name, e.BasedOn, instrument, cost))
	}

	promptInstrument := rig.Instrument
	if promptInstrument == "" {
		promptInstrument = defaultInstrument
	}
	promptInstrument = normalizeInstrument(promptInstrument, defaultInstrument)

	// 2. Hardware Capabilities
	isDualDSP := strings.Contains(hardware, "Floor") || strings.Contains(hardware, "LT") || strings.Contains(hardware, "Rack")
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
	INSTRUMENT CONTEXT: %s
	
	AVAILABLE MODELS:
	%s
	
	CONVERSATION LOGIC:
	- If the user provides feedback, adjust the technical implementation.
	- STABILITY RULE: YOU MUST STAY CONSISTENT with your previous model choices (found in conversation history). If you chose "Scream 808" previously, DO NOT change it to "Kinky Boost" in the next turn unless specifically asked to change that block.
	- You MUST stick to the Gear list proposed by the Sound Engineer.
	
	DSP MANAGEMENT:
	- CRITICAL: Physical Helix Units (Floor/Rack) have strict DSP limits.
	- Each path (Path 1 and Path 2) has its own 100%% budget.
	- YOU MUST aim for a MAXIMUM of 60-65%% per path to ensure hardware stability. 
	- If your estimated DSP sum for Path 1 exceeds 60%%, you MUST move the remaining blocks to Path 2 ("path": 1).
	- High-end Amps, Cabs, and IRs take ~30-40%% each. Poly-FX and Stereo Reverbs/Delays take ~15-25%%.
	- Path 1 is "path": 0, Path 2 is "path": 1.
	- If the hardware has only 1 path, use "path": 0 for everything.
	- BASS PRESETS: If INSTRUMENT CONTEXT is Bass, you MUST ONLY select Bass Amps and Bass Cabs from the list (e.g., SVT, Ampeg, GK, Mesa Bass). Do NOT select Guitar Amps/Cabs for Bass presets. Keep the routing simple. Avoid splitting into dual parallel chains (e.g., dual amps) unless explicitly necessary for the tone. Prefer a single chain for simplicity and DSP efficiency.

	PARAMETER CONSTRAINTS:
	- For Reverb blocks, NEVER set "Decay" or "VerbDecay" to its maximum value (1.0). Keep it at 0.7 or lower to avoid excessive noise/feedback loops.

	VARIAX AND INPUTS:
	- CRITICAL: Variax is NOT an effect block. It is a GLOBAL INPUT SETTING.
	- YOU MUST NEVER return a block named "Variax" or "Variax Simulation" in your "blocks" array.
	- THE RIG DESCRIPTION contains guitar_model and tuning fields. These are handled automatically at the input stage.
	- If the user asks for a Variax model change in a snapshot, assume it is already being handled by the Sound Engineer. Your job is ONLY to map the effect blocks in the chain.

	OUTPUT INSTRUCTIONS:
	- Return ONLY a JSON object with a "blocks" array.
	- "name" MUST match exactly the "name" of the component from the Sound Engineer proposal.
	- "model_name" must match a Name from the list.
	- "path" must be 0 (Path 1) or 1 (Path 2).
	
	OUTPUT FORMAT:
	{
		"blocks": [
			{ "name": "Tube Screamer", "model_name": "Scream 808", "path": 0, "params": { "Gain": 0.5 } }
		]
	}
	`, hardware, dspCapacity, promptInstrument, availableModels.String())

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
	bassPreset := promptInstrument == "Bass"
	useVariax := rig.UseVariax
	if rig.Instrument == "" {
		// Preserve compatibility with designs created before the resolved fields existed.
		bassPreset = isBassPreset(rig, defaultInstrument)
		useVariax = shouldApplyVariax(rig, variaxEnabled, defaultInstrument)
	}

	// 4. PRE-FLIGHT VARIAX SYNC: Ensure top-level fields are sync'd with Chain components
	// (Agents are more reliable at updating the Chain/Params than top-level technical fields)
	variaxCompName := ""
	if useVariax {
		for _, comp := range rig.Chain {
			if strings.Contains(strings.ToLower(comp.Type), "variax") || strings.Contains(strings.ToLower(comp.Name), "variax") {
				variaxCompName = comp.Name
				if rig.GuitarModel == "" || rig.GuitarModel == "None" {
					rig.GuitarModel = comp.Settings
				}
				break
			}
		}
	}
	// Sync snapshots with chain params if technical field is missing
	if variaxCompName != "" {
		for i := range rig.Snapshots {
			if rig.Snapshots[i].GuitarModel == "" || rig.Snapshots[i].GuitarModel == "None" {
				if p, ok := rig.Snapshots[i].Params[variaxCompName].(map[string]interface{}); ok {
					if m, ok := p["settings"].(string); ok {
						rig.Snapshots[i].GuitarModel = m
					} else if m, ok := p["Settings"].(string); ok {
						rig.Snapshots[i].GuitarModel = m
					}
				}
			}
		}
	}

	// 5. Parse Response
	type BuilderBlock struct {
		Name      string                 `json:"name"`
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

	// Helper to get controller map for a specific path
	getController := func(path int) map[string]interface{} {
		key := fmt.Sprintf("dsp%d", path)
		if data, ok := (*preset)["data"].(map[string]interface{}); ok {
			if tone, ok := data["tone"].(map[string]interface{}); ok {
				if ctrl, ok := tone["controller"].(map[string]interface{}); ok {
					if d, ok := ctrl[key].(map[string]interface{}); ok {
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
	hasAmp := false
	hasCab := false

	for _, b := range builderResp.Blocks {
		// VIRTUAL BLOCK SKIP: Variax is handled globally via global/snapshot logic
		if strings.Contains(strings.ToLower(b.Name), "variax") || strings.Contains(strings.ToLower(b.ModelName), "variax") {
			continue
		}

		entry, found := helix.DB.FindByRealName(b.ModelName)
		var internalID string
		var defaultData map[string]interface{}

		if found {
			internalID = entry.InternalName
			defaultData = entry.Data["Defaults"].(map[string]interface{})
		} else {
			entryID, foundID := helix.DB.FindByID(b.ModelName)
			if foundID {
				entry = entryID
				internalID = entryID.InternalName
				defaultData = entryID.Data["Defaults"].(map[string]interface{})
			} else {
				if !bassPreset {
					continue
				}
				fallback, ok := bassFallbackForRequest(b.ModelName, b.Name)
				if !ok {
					continue
				}
				entry = fallback
				internalID = fallback.InternalName
				defaultData = fallback.Data["Defaults"].(map[string]interface{})
			}
		}
		if bassPreset {
			entry = bassSafeModel(entry)
			internalID = entry.InternalName
			defaultData = entry.Data["Defaults"].(map[string]interface{})
		}
		modelType := strings.ToLower(internalID)
		hasAmp = hasAmp || strings.Contains(modelType, "_amp") || strings.Contains(modelType, "_preamp")
		hasCab = hasCab || strings.Contains(modelType, "_cab")

		finalParams := make(map[string]interface{})
		for k, v := range defaultData {
			finalParams[k] = v
		}
		finalParams["@name"] = b.Name
		finalParams["@model"] = internalID
		finalParams["@enabled"] = true

		// GET REFERENCE to clean name based on model
		technicalName := strings.ReplaceAll(strings.ReplaceAll(internalID, "HD2_", ""), "VIC_", "")
		technicalName = strings.ReplaceAll(technicalName, "_", " ")

		// Target Path (DSP Instance)
		targetPath := b.Path
		if !isDualDSP {
			targetPath = 0
		}
		// Sub-path (A/B) is locked to A (0) to avoid unnecessary split blocks
		subPath := 0

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
			finalParams["@type"] = 2
		case strings.HasPrefix(internalID, "HD2_Delay") || strings.HasPrefix(internalID, "HD2_Reverb") || strings.HasPrefix(internalID, "VIC_Reverb"):
			finalParams["@type"] = 7
		default:
			finalParams["@type"] = 0
		}

		for pName, pVal := range b.Params {
			key, ok := resolveParamKey(defaultData, pName)
			if !ok {
				// Unknown keys are not valid Helix parameters and can make HX Edit
				// or the hardware reject the whole preset.
				continue
			}
			pVal = clampParam(entry, key, normalizeParamType(entry, key, sanitizeParam(internalID, key, pVal)))
			finalParams[key] = pVal
		}

		// ENSURE ROUTING: Force sub-path 0 after AI params loop to prevent AI overwrites
		finalParams["@path"] = subPath

		// Place in correct DSP
		dsp := getDSP(targetPath)
		if dsp != nil {
			// Placing block in the correct DSP map
			blockKey := fmt.Sprintf("block%d", pos)
			dsp[blockKey] = finalParams

			// SYNC FIX: If we are processing Snapshot 0, update the main block's enabled state
			isUsedInAnySnapshot := false
			if len(rig.Snapshots) > 0 {
				for sIdx := range rig.Snapshots {
					snapshot := &rig.Snapshots[sIdx]
					for _, activeName := range snapshot.ActiveBlocks {
						an := strings.ToLower(activeName)
						bn := strings.ToLower(b.Name)
						if an == bn || strings.Contains(an, bn) || strings.Contains(bn, an) {
							isUsedInAnySnapshot = true
							break
						}
					}
					if isUsedInAnySnapshot {
						break
					}
				}

				// If block is totally unused, force it enabled in the first snapshot
				if !isUsedInAnySnapshot && len(rig.Snapshots) > 0 {
					rig.Snapshots[0].ActiveBlocks = append(rig.Snapshots[0].ActiveBlocks, b.Name)
				}
			}

			// Default Expression Pedal Assignment
			if defaultExp > 0 {
				isWah := strings.HasPrefix(internalID, "HD2_Wah")
				isVol := strings.HasPrefix(internalID, "HD2_Vol")
				isPitch := strings.HasPrefix(internalID, "HD2_PitchPitchWham")
				if isWah || isVol || isPitch {
					ctrls := getController(targetPath)
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
					if len(rig.Snapshots) > 0 {
						// Logic for custom snapshots
						for s := 0; s < 8; s++ {
							snapKey := fmt.Sprintf("snapshot%d", s)
							if snap, ok := tone[snapKey].(map[string]interface{}); ok {
								// Set snapshot name from AI proposal (up to first 4 slots)
								if s < len(rig.Snapshots) {
									snap["@name"] = rig.Snapshots[s].Name
									snap["@custom_name"] = true
								}

								// 1. Handle Bypass states
								if blocks, ok := snap["blocks"].(map[string]interface{}); ok {
									dspKey := fmt.Sprintf("dsp%d", targetPath)
									// INITIALIZATION FIX: Ensure dsp map exists for Path 2 blocks
									if _, exists := blocks[dspKey]; !exists {
										blocks[dspKey] = make(map[string]interface{})
									}

									if targetPathMap, ok := blocks[dspKey].(map[string]interface{}); ok {
										// Enable block only if it's in the snapshot's active_blocks
										isEnabled := false
										if s < len(rig.Snapshots) {
											for _, activeName := range rig.Snapshots[s].ActiveBlocks {
												a := strings.ToLower(activeName)
												// Check against both user name and technical model name for robustness
												bName := strings.ToLower(b.Name)
												bModel := strings.ToLower(b.ModelName)
												if a == bName || a == bModel || strings.Contains(bName, a) || strings.Contains(a, bName) || strings.Contains(bModel, a) || strings.Contains(a, bModel) {
													isEnabled = true
													break
												}
											}
										}

										// SYNC FIX: If we are processing Snapshot 0, update the main block's enabled state
										if s == 0 {
											if bMap, ok := dsp[blockKey].(map[string]interface{}); ok {
												bMap["@enabled"] = isEnabled
											}
										}

										targetPathMap[blockKey] = isEnabled
									}
								}

								// 2. Handle Parameter Overrides
								if s < len(rig.Snapshots) && rig.Snapshots[s].Params != nil {
									snapshot := rig.Snapshots[s]
									if overrides, ok := snapshot.Params[b.Name].(map[string]interface{}); ok {
										for pName, pVal := range overrides {
											pKey, ok := resolveParamKey(defaultData, pName)
											if !ok {
												continue
											}
											pVal = clampParam(entry, pKey, normalizeParamType(entry, pKey, sanitizeParam(internalID, pKey, pVal)))

											// OPTIMIZATION: Check if this parameter actually varies across any snapshot or from baseline
											// Only add controller if it actually changes something.
											shouldControl := false
											baselineVal := finalParams[pKey]
											for _, otherSnap := range rig.Snapshots {
												if otherOverrides, ok := otherSnap.Params[b.Name].(map[string]interface{}); ok {
													if otherVal, ok := otherOverrides[pName]; ok {
														// Sanitize otherVal to match the scale of baselineVal
														sOtherVal := sanitizeParam(internalID, pKey, otherVal)
														if sOtherVal != baselineVal {
															shouldControl = true
															break
														}
													}
												}
											}

											if shouldControl {
												// Apply snapshot controller global map
												ctrls := getController(targetPath)
												if ctrls != nil {
													if _, ok := ctrls[blockKey]; !ok {
														ctrls[blockKey] = make(map[string]interface{})
													}
													if bCtrls, ok := ctrls[blockKey].(map[string]interface{}); ok {
														bCtrls[pKey] = snapshotController(entry, pKey)
													}
												}
												// Set snapshot-specific value
												if snapCtrls, ok := snap["controllers"].(map[string]interface{}); ok {
													dspKey := fmt.Sprintf("dsp%d", targetPath)
													if _, ok := snapCtrls[dspKey]; !ok {
														snapCtrls[dspKey] = make(map[string]interface{})
													}
													if sDsp, ok := snapCtrls[dspKey].(map[string]interface{}); ok {
														if _, ok := sDsp[blockKey]; !ok {
															sDsp[blockKey] = make(map[string]interface{})
														}
														if sBlock, ok := sDsp[blockKey].(map[string]interface{}); ok {
															sBlock[pKey] = map[string]interface{}{
																"@value": pVal,
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					} else {
						// Default logic (Legacy/No Snapshots): Enable in all 8
						for s := 0; s < 8; s++ {
							snapKey := fmt.Sprintf("snapshot%d", s)
							if snap, ok := tone[snapKey].(map[string]interface{}); ok {
								if blocks, ok := snap["blocks"].(map[string]interface{}); ok {
									dspKey := fmt.Sprintf("dsp%d", targetPath)
									if d, ok := blocks[dspKey].(map[string]interface{}); ok {
										d[blockKey] = true
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if bassPreset {
		if !hasAmp {
			amp := bassFallbackEntry(false)
			addBassFallbackBlock(preset, getDSP(0), amp, amp.Name, 1, path0Count)
			path0Count++
		}
		if !hasCab {
			cab := bassFallbackEntry(true)
			addBassFallbackBlock(preset, getDSP(0), cab, cab.Name, 2, path0Count)
			path0Count++
		}
	}

	// 5. Detect Variax Intent and Apply
	variaxRequested := useVariax

	if variaxRequested {
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
		// Hardware-specific Device ID mapping
		deviceID := 15 // Default (Native)
		if strings.Contains(hardware, "Floor") || strings.Contains(hardware, "Rack") {
			deviceID = 2
		} else if strings.Contains(hardware, "LT") {
			deviceID = 3
		} else if strings.Contains(hardware, "Stomp") {
			deviceID = 6
		}
		data["@device"] = deviceID
		data["@schema"] = 0

		// EXPOSE DSP MAP: Include model->DSP costs for UI visualization
		dspMap := make(map[string]float64)
		for _, e := range helix.DB.Entries {
			cost := e.DSPMono
			if cost == 0 {
				cost = 3.0
			}
			dspMap[e.InternalName] = cost
		}
		if meta, ok := data["meta"].(map[string]interface{}); ok {
			meta["dsp_map"] = dspMap
		}

		if tone, ok := data["tone"].(map[string]interface{}); ok {
			// HARDWARE OUTPUTS: Force Multi Output (1) by default instead of Native-Host (15)
			if dsp0, ok := tone["dsp0"].(map[string]interface{}); ok {
				if outA, ok := dsp0["outputA"].(map[string]interface{}); ok {
					outA["@output"] = 1 // Default Multi
				}
			}
			if dsp1, ok := tone["dsp1"].(map[string]interface{}); ok {
				if outA, ok := dsp1["outputA"].(map[string]interface{}); ok {
					outA["@output"] = 1 // Default Multi
				}
			}

			// SERIAL INTERCONNECT: Route DSP0 to Path 2 only if Path 2 is actually used
			if isDualDSP && path1Count > 0 {
				if dsp0, ok := tone["dsp0"].(map[string]interface{}); ok {
					if outA, ok := dsp0["outputA"].(map[string]interface{}); ok {
						outA["@output"] = 2 // Explicit "Path 2" output for serially linking DSPs
					}
				}
			}

			if global, ok := tone["global"].(map[string]interface{}); ok {
				global["@cursor_dsp"] = 0
				global["@cursor_group"] = "inputA"
			}
		}
	}

	return preset, nil
}

// isBassInstrument identifies bass defaults without treating unrelated guitar descriptions as bass intent.
func isBassInstrument(instrument string) bool {
	return strings.Contains(strings.ToLower(instrument), "bass")
}

// isBassPreset preserves legacy bass detection for designs created before instrument resolution was added.
func isBassPreset(rig *RigDescription, instrument string) bool {
	model := strings.ToLower(rig.GuitarModel)
	if isBassInstrument(model) || strings.Contains(model, "stingray") || strings.Contains(model, "music man") {
		return true
	}
	if rig.GuitarModel != "" && rig.GuitarModel != "None" {
		return false
	}
	for _, component := range rig.Chain {
		value := strings.ToLower(component.Name + " " + component.Description + " " + component.Settings)
		if strings.Contains(value, "variax") {
			return false
		}
		if strings.Contains(value, "bass") || strings.Contains(value, "gallien") || strings.Contains(value, "ampeg") || strings.Contains(value, "stingray") {
			return true
		}
	}
	return isBassInstrument(instrument)
}

// shouldApplyVariax applies the legacy Variax decision while preventing Variax use for bass designs.
func shouldApplyVariax(rig *RigDescription, enabled bool, instrument string) bool {
	if isBassPreset(rig, instrument) {
		return false
	}
	if enabled || (rig.GuitarModel != "" && rig.GuitarModel != "None") {
		return true
	}
	for _, snapshot := range rig.Snapshots {
		if snapshot.GuitarModel != "" && snapshot.GuitarModel != "None" {
			return true
		}
	}
	return false
}

// bassSafeModel replaces a guitar-only amp or cab with a catalogued bass-compatible fallback.
func bassSafeModel(entry helix.CatalogEntry) helix.CatalogEntry {
	modelType := strings.ToLower(entry.InternalName)
	isAmp := strings.Contains(modelType, "_amp") || strings.Contains(modelType, "_preamp")
	isCab := strings.Contains(modelType, "_cab")
	if !isAmp && !isCab || isBassCompatibleModel(entry) {
		return entry
	}

	return bassFallbackEntry(isCab)
}

// bassFallbackForRequest maps an unresolved bass block request to a safe catalog entry.
func bassFallbackForRequest(modelName, blockName string) (helix.CatalogEntry, bool) {
	requested := strings.ToLower(modelName + " " + blockName)
	isCab := false
	if strings.Contains(requested, "cab") || strings.Contains(requested, "micir") ||
		strings.Contains(requested, "1x10") || strings.Contains(requested, "1x15") ||
		strings.Contains(requested, "2x15") || strings.Contains(requested, "4x10") ||
		strings.Contains(requested, "6x10") || strings.Contains(requested, "8x10") {
		isCab = true
	}
	if isCab {
		return bassFallbackEntry(true), true
	}
	if strings.Contains(requested, "svt") || strings.Contains(requested, "ampeg") {
		return bassFallbackEntry(false), true
	}
	if strings.Contains(requested, "cougar") || strings.Contains(requested, "gallien") {
		entry, ok := helix.DB.FindByID("HD2_AmpGCougar800")
		return entry, ok
	}
	return bassFallbackEntry(false), true
}

// bassFallbackEntry returns the standard bass amp or cab used when generation omits one.
func bassFallbackEntry(isCab bool) helix.CatalogEntry {
	if isCab {
		entry, _ := helix.DB.FindByID("HD2_CabMicIr_8x10SVTAV")
		return entry
	}
	entry, _ := helix.DB.FindByID("HD2_AmpSVBeastNrm")
	return entry
}

// addBassFallbackBlock inserts a generated bass block and enables it in every snapshot.
func addBassFallbackBlock(preset *helix.Preset, dsp map[string]interface{}, entry helix.CatalogEntry, name string, blockType, position int) {
	if dsp == nil {
		return
	}
	defaults, ok := entry.Data["Defaults"].(map[string]interface{})
	if !ok {
		return
	}
	params := make(map[string]interface{}, len(defaults)+6)
	for key, value := range defaults {
		params[key] = value
	}
	params["@name"] = name
	params["@model"] = entry.InternalName
	params["@type"] = blockType
	params["@enabled"] = true
	params["@position"] = position
	params["@path"] = 0
	blockKey := fmt.Sprintf("block%d", position)
	dsp[blockKey] = params

	data, ok := (*preset)["data"].(map[string]interface{})
	if !ok {
		return
	}
	tone, ok := data["tone"].(map[string]interface{})
	if !ok {
		return
	}
	for snapshotIndex := 0; snapshotIndex < 8; snapshotIndex++ {
		snapshot, ok := tone[fmt.Sprintf("snapshot%d", snapshotIndex)].(map[string]interface{})
		if !ok {
			continue
		}
		blocks, ok := snapshot["blocks"].(map[string]interface{})
		if !ok {
			blocks = make(map[string]interface{})
			snapshot["blocks"] = blocks
		}
		dspBlocks, ok := blocks["dsp0"].(map[string]interface{})
		if !ok {
			dspBlocks = make(map[string]interface{})
			blocks["dsp0"] = dspBlocks
		}
		dspBlocks[blockKey] = true
	}
}

// isBassCompatibleModel accepts catalog entries marked as bass or known bass hardware models.
func isBassCompatibleModel(entry helix.CatalogEntry) bool {
	if entry.Instrument == "Bass" {
		return true
	}
	model := strings.ToLower(entry.Name + " " + entry.BasedOn + " " + entry.InternalName)
	for _, modelID := range []string{
		"ampbusyone", "ampcalibass", "ampcali400", "ampmandarinbass", "ampwoodyblue", "ampaguasledge", "ampagua51", "ampdelsol",
		"cabmicir_1x12epicenter", "cabmicir_1x15ampeg", "cabmicir_2x15", "cabmicir_4x10garden", "cabmicir_4x10ampeg", "cabmicir_6x10cali", "cabmicir_8x10svt",
		"cab1x12delsol", "cab1x15tuckngo", "cab1x18delsol", "cab1x18woody", "cab2x15brute", "cab4x10rhino", "cab6x10cali", "cab8x10svbeast",
	} {
		if strings.Contains(model, modelID) {
			return true
		}
	}
	for _, keyword := range []string{"bass", "svt", "ampeg", "aguilar", "gallien", "acoustic 360"} {
		if strings.Contains(model, keyword) {
			return true
		}
	}
	return false
}

// applyVariax configures global and per-snapshot Variax models, tuning, controller metadata, and hardware type for a preset.
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

	// 0. Load Configuration Once
	type ConfigRoot struct {
		Configs map[string]struct {
			Inherits     string `json:"inherits"`
			VariantLogic string `json:"variant_logic"`
			Banks        []struct {
				Name   string `json:"name"`
				BaseID int    `json:"base_id"`
			} `json:"banks"`
			Aliases map[string]string `json:"aliases"`
			Tunings map[string]struct {
				Offsets []int    `json:"offsets"`
				Aliases []string `json:"aliases"`
			} `json:"tunings"`
		} `json:"variax_configurations"`
	}

	confPath := "pkg/helix/data/variax_models.json"
	configBytes, _ := os.ReadFile(confPath)
	var root ConfigRoot
	if configBytes != nil {
		json.Unmarshal(configBytes, &root)
	}

	// Internal helper to get active config
	getActiveCfg := func(hw string) *ConfigRoot {
		return &root // keeping root reference for ease, though we could resolve hw here
		// For simplicity in this logic, helpers will resolve their own 'cfg' from 'root'
	}
	_ = getActiveCfg // avoid unused

	// Helper to map model name (Hardware-Aware JSON Driven)
	mapModel := func(modelName string, hw string) int {
		m := strings.ToLower(modelName)
		if m == "" || m == "none" {
			return -1
		}
		if m == "0" || m == "neutral" {
			return 0
		}

		hwKey := strings.ToLower(hw)
		cfg, ok := root.Configs[hwKey]
		if !ok {
			for k, v := range root.Configs {
				if strings.Contains(hwKey, k) {
					cfg = v
					ok = true
					break
				}
			}
		}
		if ok && cfg.Inherits != "" {
			cfg = root.Configs[cfg.Inherits]
		}

		getFallback := func() int {
			if strings.Contains(m, "jaguar") || strings.Contains(m, "tele") || strings.Contains(m, "t-model") {
				return 10
			}
			if strings.Contains(m, "strat") || strings.Contains(m, "spank") {
				return 15
			}
			return -1
		}
		if !ok {
			return getFallback()
		}

		targetBank := ""
		for alias, bank := range cfg.Aliases {
			if strings.Contains(m, alias) {
				targetBank = bank
				break
			}
		}
		var baseID int = -1
		for _, b := range cfg.Banks {
			if strings.EqualFold(b.Name, targetBank) || strings.Contains(m, strings.ToLower(b.Name)) {
				baseID = b.BaseID
				break
			}
		}
		if baseID >= 0 {
			variant := 1
			// Find first digit 1-5 in the entire string (e.g. from "(Pickup Pos 2)")
			for _, char := range m {
				if char >= '1' && char <= '5' {
					variant = int(char - '0')
					break
				}
			}

			if cfg.VariantLogic == "inverted" {
				return baseID + (5 - variant)
			}
			return baseID + (variant - 1)
		}
		return getFallback()
	}

	// Helper to get tuning offsets (Hardware-Aware JSON Driven)
	getTuningOffsets := func(tuning string, hw string) ([6]int, bool) {
		t := strings.ToLower(tuning)
		if t == "" || t == "standard" {
			return [6]int{0, 0, 0, 0, 0, 0}, false
		}

		hwKey := strings.ToLower(hw)
		cfg, ok := root.Configs[hwKey]
		if !ok {
			for k, v := range root.Configs {
				if strings.Contains(hwKey, k) {
					cfg = v
					ok = true
					break
				}
			}
		}
		if ok && cfg.Inherits != "" {
			cfg = root.Configs[cfg.Inherits]
		}

		// 1. Config Match
		if ok {
			for name, data := range cfg.Tunings {
				if strings.EqualFold(name, tuning) {
					var off [6]int
					copy(off[:], data.Offsets)
					return off, true
				}
				for _, alias := range data.Aliases {
					if strings.EqualFold(alias, tuning) || strings.Contains(t, strings.ToLower(alias)) {
						var off [6]int
						copy(off[:], data.Offsets)
						return off, true
					}
				}
			}
		}

		// 2. Keyword Fallback
		switch {
		case strings.Contains(t, "drop d"):
			return [6]int{-2, 0, 0, 0, 0, 0}, true
		case strings.Contains(t, "eb") || strings.Contains(t, "half step down"):
			return [6]int{-1, -1, -1, -1, -1, -1}, true
		case strings.Contains(t, "d standard") || strings.Contains(t, "whole step down"):
			return [6]int{-2, -2, -2, -2, -2, -2}, true
		case strings.Contains(t, "drop c"):
			return [6]int{-4, -2, -2, -2, -2, -2}, true
		case strings.Contains(t, "baritone"):
			return [6]int{-5, -5, -5, -5, -5, -5}, true
		case strings.Contains(t, "open g"):
			return [6]int{-2, -2, 0, 0, 0, -2}, true
		case strings.Contains(t, "open d"):
			return [6]int{-2, 0, 0, -1, -2, -2}, true
		case strings.Contains(t, "dadgad"):
			return [6]int{-2, 0, 0, 0, 0, -2}, true
		}
		return [6]int{0, 0, 0, 0, 0, 0}, false
	}

	// 1. Global Model Selection
	modelID := mapModel(rig.GuitarModel, hardwareModel)
	if modelID >= 0 {
		v["@variax_model"] = modelID
	}
	v["@variax_magmode"] = true

	// Expose Variax Type to UI meta
	if meta, ok := data["meta"].(map[string]interface{}); ok {
		vType := "jtv"
		if strings.Contains(strings.ToLower(hardwareModel), "shuriken") {
			vType = "shuriken"
		}
		meta["variax_type"] = vType
	}

	// 2. Global Tuning Logic
	offsets, isTuningMapped := getTuningOffsets(rig.Tuning, hardwareModel)
	if isTuningMapped {
		v["@variax_customtuning"] = true
		v["@variax_str1tuning"] = offsets[5]
		v["@variax_str2tuning"] = offsets[4]
		v["@variax_str3tuning"] = offsets[3]
		v["@variax_str4tuning"] = offsets[2]
		v["@variax_str5tuning"] = offsets[1]
		v["@variax_str6tuning"] = offsets[0]
	} else {
		v["@variax_customtuning"] = false
	}

	// 3. Register Variax Controller (Enables Snapshot Control on Hardware)
	if ctrl, ok := tone["controller"].(map[string]interface{}); ok {
		if _, ok := ctrl["variax"]; !ok {
			ctrl["variax"] = make(map[string]interface{})
		}
		if cv, ok := ctrl["variax"].(map[string]interface{}); ok {
			cv["@variax_model"] = map[string]interface{}{
				"@controller":       19, // Snapshot Control
				"@globalblock":      "inputA",
				"@globaldsp":        0,
				"@max":              60,
				"@min":              0,
				"@snapshot_disable": false,
			}
		}
	}

	// 4. Per-Snapshot Overrides (Standard Hardware Format)
	if len(rig.Snapshots) > 0 {
		for s := 0; s < 8; s++ {
			snapKey := fmt.Sprintf("snapshot%d", s)
			if snap, ok := tone[snapKey].(map[string]interface{}); ok {
				// Initialize snapshot controllers if missing
				if _, ok := snap["controllers"]; !ok {
					snap["controllers"] = make(map[string]interface{})
				}
				sCtrls := snap["controllers"].(map[string]interface{})
				if _, ok := sCtrls["variax"]; !ok {
					sCtrls["variax"] = make(map[string]interface{})
				}
				vCtrl := sCtrls["variax"].(map[string]interface{})

				// Fetch current model (default to global if not specified for this snapshot)
				currentModelID := modelID
				sModelStr := ""
				if s < len(rig.Snapshots) {
					sModelStr = rig.Snapshots[s].GuitarModel
				}

				// FALLBACK: Search snapshot params for metadata if field is empty
				if s < len(rig.Snapshots) && (sModelStr == "" || sModelStr == "None") {
					if rig.Snapshots[s].Params != nil {
						for bName, bParams := range rig.Snapshots[s].Params {
							if strings.Contains(strings.ToLower(bName), "variax") {
								if paramsMap, ok := bParams.(map[string]interface{}); ok {
									// Sound Engineer often puts it in a 'Model' or 'Settings' key inside params
									if m, ok := paramsMap["Model"].(string); ok {
										sModelStr = m
									} else if m, ok := paramsMap["Settings"].(string); ok {
										sModelStr = m
									}
								}
							}
						}
					}
				}

				if sModelStr != "" && sModelStr != "None" {
					sModelID := mapModel(sModelStr, hardwareModel)
					if sModelID >= 0 {
						currentModelID = sModelID
					}
				}

				// Apply model to snapshot controller
				if currentModelID >= 0 {
					vCtrl["@variax_model"] = map[string]interface{}{
						"@fs_enabled": false,
						"@value":      currentModelID,
					}
				}
			}
		}
	}
}

// It matches exact names, known aliases, and case-insensitive equivalents.
func resolveParamKey(defaults map[string]interface{}, requested string) (string, bool) {
	if _, ok := defaults[requested]; ok {
		return requested, true
	}

	aliases := map[string][]string{
		"gain":   {"Drive", "LeadGain", "Lead Drive", "ChVol", "Master"},
		"drive":  {"Gain", "LeadDrive", "Lead Gain", "Overdrive"},
		"volume": {"ChVol", "Master", "Level"},
		"vol":    {"ChVol", "Master", "Level"},
		"mids":   {"Middle", "Mid"},
		"31hz":   {"31p25Hz"},
		"62hz":   {"62p5Hz"},
	}
	if candidates, ok := aliases[strings.ToLower(requested)]; ok {
		for _, candidate := range candidates {
			if _, exists := defaults[candidate]; exists {
				return candidate, true
			}
		}
	}

	for actual := range defaults {
		if strings.EqualFold(actual, requested) {
			return actual, true
		}
	}
	return "", false
}

// parameterLimits returns the minimum and maximum hardware limits for a catalog parameter.
// The boolean reports whether both limits are available.
func parameterLimits(entry helix.CatalogEntry, key string) (float64, float64, bool) {
	dict, ok := entry.Data["Controller_Dict"].(map[string]interface{})
	if !ok {
		return 0, 0, false
	}
	spec, ok := dict[key].(map[string]interface{})
	if !ok {
		return 0, 0, false
	}
	min, minOK := spec["@min"].(float64)
	max, maxOK := spec["@max"].(float64)
	return min, max, minOK && maxOK
}

// clampParam restricts a numeric parameter to the limits defined in the hardware catalog.
// Values that are not numeric or have no catalog limits are returned unchanged.
func clampParam(entry helix.CatalogEntry, key string, value interface{}) interface{} {
	number, ok := value.(float64)
	if !ok {
		return value
	}
	min, max, ok := parameterLimits(entry, key)
	if !ok {
		return value
	}
	if number < min {
		return min
	}
	if number > max {
		return max
	}
	return number
}

// normalizeParamType falls back to the catalog default when an AI value has the wrong type.
func normalizeParamType(entry helix.CatalogEntry, key string, value interface{}) interface{} {
	defaults, ok := entry.Data["Defaults"].(map[string]interface{})
	if !ok {
		return value
	}
	defaultValue, ok := defaults[key]
	if !ok {
		return value
	}

	switch defaultValue.(type) {
	case bool:
		if _, ok := value.(bool); !ok {
			return defaultValue
		}
	case string:
		if _, ok := value.(string); !ok {
			return defaultValue
		}
	case float64:
		if _, ok := value.(float64); !ok {
			return defaultValue
		}
	}
	return value
}

// snapshotController creates a snapshot controller definition using the parameter's catalog limits when available.
func snapshotController(entry helix.CatalogEntry, key string) map[string]interface{} {
	min, max := 0.0, 1.0
	if catalogMin, catalogMax, ok := parameterLimits(entry, key); ok {
		min, max = catalogMin, catalogMax
	}
	return map[string]interface{}{
		"@controller":       19,
		"@max":              max,
		"@min":              min,
		"@snapshot_disable": false,
	}
}

// sanitizeParam applies safety normalization and limits to numeric effect parameters before catalog validation. Reverb decay is capped at 0.7, and delay feedback is capped at 0.75.
func sanitizeParam(internalID, k string, v interface{}) interface{} {
	val, isFloat := v.(float64)
	if !isFloat {
		return v
	}

	lowK := strings.ToLower(k)

	// List of parameters that should be in 0.0-1.0 range internally (0-10 on knob)
	isPercentageParam := strings.Contains(lowK, "gain") ||
		strings.Contains(lowK, "drive") ||
		strings.Contains(lowK, "bass") ||
		strings.Contains(lowK, "mid") ||
		strings.Contains(lowK, "treble") ||
		strings.Contains(lowK, "presence") ||
		strings.Contains(lowK, "chvol") ||
		strings.Contains(lowK, "master") ||
		strings.Contains(lowK, "level") ||
		strings.Contains(lowK, "mix") ||
		strings.Contains(lowK, "feedback") ||
		strings.Contains(lowK, "fdbk") ||
		strings.Contains(lowK, "pedal")

	if isPercentageParam {
		// Normalization: If the AI provides 3.5, it likely meant 0.35
		if val > 1.0 {
			val = val / 10.0
		}
		// Hard Cap: Ensure we don't exceed 1.0 for these parameters
		if val > 1.0 {
			val = 1.0
		}
		v = val
	}

	// Reverb Decay Sanitization (Reverbs and Delays with Reverb tails)
	isReverb := strings.HasPrefix(internalID, "HD2_Reverb") || strings.HasPrefix(internalID, "VIC_Reverb")
	isDelay := strings.HasPrefix(internalID, "HD2_Delay") || strings.HasPrefix(internalID, "VIC_Delay")
	if (isReverb && (lowK == "decay" || lowK == "verbdecay")) || (isDelay && lowK == "verbdecay") {
		if val >= 0.7 {
			return 0.7
		}
	}

	// Delay Feedback Sanitization
	if isDelay && (lowK == "feedback" || lowK == "fdbk" || lowK == "bk") {
		if val >= 0.75 {
			return 0.75
		}
	}

	return v
}
