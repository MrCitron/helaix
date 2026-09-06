package gemini

import (
	"HelAIx/pkg/helix"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"google.golang.org/genai"
)

type builderBlock struct {
	Name      string                 `json:"name"`
	ModelName string                 `json:"model_name"`
	Path      int                    `json:"path"`
	Params    map[string]interface{} `json:"params"`
}

type builderResponse struct {
	Blocks []builderBlock `json:"blocks"`
}

const plannerBeamWidth = 256

type rankedComponentCandidates struct {
	component  RigComponent
	candidates []helix.CatalogEntry
}

type plannedBlock struct {
	name      string
	modelName string
	path      int
	dsp       float64
}

type recommendedCandidatePlan struct {
	blocks []plannedBlock
	dsp    [2]float64
}

// ChatPresetEngineer takes the abstract rig and maps it to specific Helix Blocks, or refines an existing implementation
func (c *Client) ChatPresetEngineer(ctx context.Context, rig *RigDescription, presetName string, history []ChatMessage, hardware string, defaultExp int, variaxEnabled bool, hardwareModel string) (*helix.Preset, error) {
	// 1. Prepare only the models that can implement each proposed component.
	availableModels, err := CandidateCatalogForRig(*rig)
	if err != nil {
		return nil, err
	}

	// 2. Hardware Capabilities
	isDualDSP := IsDualDSPHardware(hardware)
	recommendedPlan, err := RecommendedCandidatePlanForRig(*rig, isDualDSP)
	if err != nil {
		return nil, err
	}
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
	
	TOP-RANKED ALLOWED MODELS BY COMPONENT:
	%s

	%s

	PLANNER RULES:
	- The recommended plan is a validated combination, not independent suggestions. Copy every name, model_name, and path tuple exactly unless the user explicitly asks to change that component.
	- Never substitute a model or move a block to another path while retaining the rest of the plan without rechecking the full chain against the allowed candidates and DSP budget.
	- Return every non-Variax component exactly once, in the rig's chain order. When a dual-DSP plan moves to path 1, do not place later chain blocks back on path 0.
	
	CONVERSATION LOGIC:
	- If the user provides feedback, adjust the technical implementation.
	- STABILITY RULE: YOU MUST STAY CONSISTENT with your previous model choices (found in conversation history). If you chose "Scream 808" previously, DO NOT change it to "Kinky Boost" in the next turn unless specifically asked to change that block.
	- You MUST stick to the Gear list proposed by the Sound Engineer.
	
	DSP MANAGEMENT:
	- CRITICAL: Physical Helix Units (Floor/Rack) have strict DSP limits.
	- Each path (Path 1 and Path 2) has its own 100%% budget.
	- The recommended plan is already within the deterministic 65%% per-path budget. Do not estimate or improvise an alternative path allocation.
	- If user feedback requires a different model, select only an allowed candidate for that component and keep each path at or below 65%%.
	- High-end Amps, Cabs, and IRs take ~30-40%% each. Poly-FX and Stereo Reverbs/Delays take ~15-25%%.
	- Path 1 is "path": 0, Path 2 is "path": 1.
	- If the hardware has only 1 path, use "path": 0 for everything.

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
	- "model_name" must match a candidate listed for that exact component; candidates are already ranked for the requested tone.
	- Use the recommended DSP-safe plan as the exact baseline unless user feedback explicitly requires a change.
	- "path" must be 0 (Path 1) or 1 (Path 2).
	
	OUTPUT FORMAT:
	{
		"blocks": [
			{ "name": "Tube Screamer", "model_name": "Scream 808", "path": 0, "params": { "Gain": 0.5 } }
		]
	}
	`, hardware, dspCapacity, availableModels, recommendedPlan)

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
	preset, buildErr := BuildPresetFromJSON(jsonText, rig, presetName, hardware, defaultExp, variaxEnabled, hardwareModel)
	if buildErr == nil {
		return preset, nil
	}

	contents = append(contents,
		&genai.Content{Role: "model", Parts: []*genai.Part{{Text: jsonText}}},
		&genai.Content{Role: "user", Parts: []*genai.Part{{Text: BuilderCorrectionInstruction(buildErr)}}},
	)
	retry, err := c.client.Models.GenerateContent(ctx, c.ModelName, contents, config)
	if err != nil {
		return nil, fmt.Errorf("preset engineer retry failed: %v", err)
	}
	if len(retry.Candidates) == 0 || len(retry.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty retry response from Preset Engineer Agent")
	}
	preset, retryErr := BuildPresetFromJSON(retry.Candidates[0].Content.Parts[0].Text, rig, presetName, hardware, defaultExp, variaxEnabled, hardwareModel)
	if retryErr != nil {
		return nil, fmt.Errorf("preset engineer response remained invalid after retry: %w", retryErr)
	}
	return preset, nil
}

// CandidateCatalogForRig formats the highest-ranked compatible Helix models for each component.
func CandidateCatalogForRig(rig RigDescription) (string, error) {
	groups, err := rankedCandidateGroups(rig)
	if err != nil {
		return "", err
	}
	var catalog strings.Builder
	for _, group := range groups {
		catalog.WriteString(fmt.Sprintf("%s [%s, ranked by tone intent]:\n", group.component.Name, group.component.Type))
		for _, candidate := range group.candidates {
			name := candidate.Name
			if name == "" {
				name = candidate.InternalName
			}
			cost := helix.EffectiveDSPMono(candidate)
			catalog.WriteString(fmt.Sprintf("- %s (Based on: %s) [DSP: %.1f%%]\n", name, candidate.BasedOn, cost))
		}
	}
	return catalog.String(), nil
}

// IsDualDSPHardware reports whether the selected Helix hardware supports two DSP paths.
func IsDualDSPHardware(hardware string) bool {
	return strings.Contains(hardware, "Floor") || strings.Contains(hardware, "LT") || strings.Contains(hardware, "Rack")
}

// RecommendedCandidatePlanForRig returns a DSP-safe, tone-ranked starting plan for builders.
func RecommendedCandidatePlanForRig(rig RigDescription, isDualDSP bool) (string, error) {
	plan, err := recommendedPlanForRig(rig, isDualDSP)
	if err != nil {
		return "", err
	}
	var text strings.Builder
	text.WriteString("RECOMMENDED DSP-SAFE STARTING PLAN:\n")
	for _, block := range plan.blocks {
		text.WriteString(fmt.Sprintf("- %s -> %s (path %d, DSP %.1f%%)\n", block.name, block.modelName, block.path, block.dsp))
	}
	text.WriteString(fmt.Sprintf("Path 0 total: %.1f%% DSP\n", plan.dsp[0]))
	if isDualDSP {
		text.WriteString(fmt.Sprintf("Path 1 total: %.1f%% DSP\n", plan.dsp[1]))
	}
	return text.String(), nil
}

// BuilderCorrectionInstruction tells a provider how to replace a rejected block mapping.
func BuilderCorrectionInstruction(validationErr error) string {
	return fmt.Sprintf(`Your previous JSON was rejected by deterministic Helix validation: %v.
Return a complete replacement JSON object only, with a full "blocks" array; do not return a patch, explanation, or extra keys.
- Include every non-Variax rig component exactly once and in chain order.
- Use only the exact allowed model names and the recommended DSP-safe model/path tuples unless the requested change requires a different allowed candidate.
- Keep each path at or below 65%% DSP. Once the chain moves to path 1, do not move later blocks back to path 0.`, validationErr)
}

func rankedCandidateGroups(rig RigDescription) ([]rankedComponentCandidates, error) {
	groups := make([]rankedComponentCandidates, 0, len(rig.Chain))
	for _, component := range rig.Chain {
		if component.Type == "variax" {
			continue
		}
		intent := strings.Join([]string{component.Name, component.Description, component.Settings}, " ")
		candidates := helix.RankedCandidatesForComponent(component.Type, intent, helix.PromptCandidateLimit)
		if len(candidates) == 0 {
			return nil, fmt.Errorf("no Helix candidates available for component %q of type %q", component.Name, component.Type)
		}
		groups = append(groups, rankedComponentCandidates{component: component, candidates: candidates})
	}
	return groups, nil
}

func recommendedPlanForRig(rig RigDescription, isDualDSP bool) (recommendedCandidatePlan, error) {
	groups, err := rankedCandidateGroups(rig)
	if err != nil {
		return recommendedCandidatePlan{}, err
	}
	return recommendedPlanForGroups(groups, isDualDSP)
}

func recommendedPlanForGroups(groups []rankedComponentCandidates, isDualDSP bool) (recommendedCandidatePlan, error) {
	type planState struct {
		plan      recommendedCandidatePlan
		rankScore int
		path      int
	}
	states := []planState{{}}
	for _, group := range groups {
		next := make([]planState, 0, len(states)*len(group.candidates)*2)
		for _, state := range states {
			for rank, candidate := range group.candidates {
				cost := helix.EffectiveDSPMono(candidate)
				for _, path := range plannerPaths(state.path, isDualDSP) {
					if state.plan.dsp[path]+cost > helix.SafeDSPPerPath {
						continue
					}
					modelName := candidate.Name
					if modelName == "" {
						modelName = candidate.InternalName
					}
					plan := state.plan
					plan.blocks = append(append([]plannedBlock(nil), state.plan.blocks...), plannedBlock{name: group.component.Name, modelName: modelName, path: path, dsp: cost})
					plan.dsp[path] += cost
					next = append(next, planState{plan: plan, rankScore: state.rankScore + rank, path: path})
				}
			}
		}
		if len(next) == 0 {
			return recommendedCandidatePlan{}, fmt.Errorf("no DSP-safe candidate plan for component %q", group.component.Name)
		}
		sort.Slice(next, func(i, j int) bool {
			if next[i].rankScore != next[j].rankScore {
				return next[i].rankScore < next[j].rankScore
			}
			leftDSP := next[i].plan.dsp[0] + next[i].plan.dsp[1]
			rightDSP := next[j].plan.dsp[0] + next[j].plan.dsp[1]
			if leftDSP != rightDSP {
				return leftDSP < rightDSP
			}
			return next[i].path < next[j].path
		})
		if len(next) > plannerBeamWidth {
			next = next[:plannerBeamWidth]
		}
		states = next
	}
	return states[0].plan, nil
}

func plannerPaths(currentPath int, isDualDSP bool) []int {
	if !isDualDSP || currentPath == 1 {
		return []int{currentPath}
	}
	return []int{0, 1}
}

// BuildPresetFromJSON validates a provider's block mapping and creates an exportable preset.
func BuildPresetFromJSON(jsonText string, rig *RigDescription, presetName string, hardware string, defaultExp int, variaxEnabled bool, hardwareModel string) (*helix.Preset, error) {
	helix.DB.EnsureLoaded()
	if rig == nil {
		return nil, fmt.Errorf("rig description is required")
	}
	if err := validateRigDescription(rig); err != nil {
		return nil, fmt.Errorf("invalid rig description: %w", err)
	}
	isDualDSP := IsDualDSPHardware(hardware)

	// 4. PRE-FLIGHT VARIAX SYNC: Ensure top-level fields are sync'd with Chain components
	// (Agents are more reliable at updating the Chain/Params than top-level technical fields)
	variaxCompName := ""
	for _, comp := range rig.Chain {
		if strings.Contains(strings.ToLower(comp.Type), "variax") || strings.Contains(strings.ToLower(comp.Name), "variax") {
			variaxCompName = comp.Name
			if rig.GuitarModel == "" || rig.GuitarModel == "None" {
				rig.GuitarModel = comp.Settings
			}
			break
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

	// 5. Parse and validate the model response before using it to build a preset.
	var builderResp builderResponse
	decoder := json.NewDecoder(bytes.NewBufferString(jsonText))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&builderResp); err != nil {
		return nil, fmt.Errorf("failed to parse Preset Engineer JSON: %v. Raw: %s", err, jsonText)
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("Preset Engineer returned multiple JSON values")
	}
	if err := validateBuilderResponse(builderResp, rig, isDualDSP); err != nil {
		return nil, fmt.Errorf("invalid Preset Engineer response: %w", err)
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

	for _, b := range builderResp.Blocks {
		// VIRTUAL BLOCK SKIP: Variax is handled globally via global/snapshot logic
		if strings.Contains(strings.ToLower(b.Name), "variax") || strings.Contains(strings.ToLower(b.ModelName), "variax") {
			continue
		}

		entry, found := helix.DB.FindByRealName(b.ModelName)
		if !found {
			entry, found = helix.DB.FindByID(b.ModelName)
		}
		if !found {
			return nil, fmt.Errorf("model %q disappeared from the Helix catalog", b.ModelName)
		}
		defaultData, ok := entry.Data["Defaults"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("model %q has invalid defaults", b.ModelName)
		}
		internalID := entry.InternalName

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
			key, err := resolveParameterName(entry, pName, false)
			if err != nil {
				return nil, fmt.Errorf("block %q: %w", b.Name, err)
			}
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
											pKey, err := resolveParameterName(entry, pName, true)
											if err != nil {
												return nil, fmt.Errorf("snapshot %q block %q: %w", snapshot.Name, b.Name, err)
											}

											// OPTIMIZATION: Check if this parameter actually varies across any snapshot or from baseline
											// Only add controller if it actually changes something.
											shouldControl := false
											baselineVal := finalParams[pKey]
											for _, otherSnap := range rig.Snapshots {
												if otherOverrides, ok := otherSnap.Params[b.Name].(map[string]interface{}); ok {
													for otherName, otherVal := range otherOverrides {
														otherKey, err := resolveParameterName(entry, otherName, true)
														if err != nil {
															return nil, fmt.Errorf("snapshot %q block %q: %w", otherSnap.Name, b.Name, err)
														}
														if otherKey == pKey && otherVal != baselineVal {
															shouldControl = true
															break
														}
													}
													if shouldControl {
														break
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
														bCtrls[pKey] = map[string]interface{}{
															"@controller":       9, // Snapshot Controller
															"@max":              1.0,
															"@min":              0.0,
															"@snapshot_disable": false,
														}
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

	// 5. Detect Variax Intent and Apply
	variaxRequested := variaxEnabled
	if rig.GuitarModel != "" && rig.GuitarModel != "None" {
		variaxRequested = true
	}
	for _, s := range rig.Snapshots {
		if s.GuitarModel != "" && s.GuitarModel != "None" {
			variaxRequested = true
			break
		}
	}

	if variaxRequested {
		if err := applyVariax(preset, rig, hardwareModel); err != nil {
			return nil, err
		}
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
			dspMap[e.InternalName] = helix.EffectiveDSPMono(e)
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

	if err := helix.ValidatePreset(*preset); err != nil {
		return nil, fmt.Errorf("generated preset failed validation: %w", err)
	}

	return preset, nil
}

func applyVariax(preset *helix.Preset, rig *RigDescription, hardwareModel string) error {
	data, ok := (*preset)["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("preset is missing data")
	}
	tone, ok := data["tone"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("preset is missing tone data")
	}
	v, ok := tone["variax"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("preset is missing variax data")
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

	var root ConfigRoot
	if err := json.Unmarshal(helix.VariaxModelsJSON, &root); err != nil {
		return fmt.Errorf("failed to load embedded Variax configuration: %w", err)
	}
	if len(root.Configs) == 0 {
		return fmt.Errorf("embedded Variax configuration is empty")
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

	return nil
}

func validateBuilderResponse(response builderResponse, rig *RigDescription, isDualDSP bool) error {
	expected := make(map[string]string)
	for _, component := range rig.Chain {
		if strings.Contains(strings.ToLower(component.Type), "variax") || strings.Contains(strings.ToLower(component.Name), "variax") {
			continue
		}
		if component.Name == "" {
			return fmt.Errorf("rig contains a component without a name")
		}
		if _, exists := expected[component.Name]; exists {
			return fmt.Errorf("rig contains duplicate component name %q", component.Name)
		}
		expected[component.Name] = component.Type
	}

	if len(expected) == 0 {
		return fmt.Errorf("rig has no mappable components")
	}
	if len(response.Blocks) != len(expected) {
		return fmt.Errorf("expected %d blocks, got %d", len(expected), len(response.Blocks))
	}

	seen := make(map[string]struct{}, len(response.Blocks))
	blockModels := make(map[string]helix.CatalogEntry, len(response.Blocks))
	pathDSP := [2]float64{}
	for _, block := range response.Blocks {
		if block.Name == "" || block.ModelName == "" {
			return fmt.Errorf("block name and model_name are required")
		}
		componentType, ok := expected[block.Name]
		if !ok {
			return fmt.Errorf("block %q is not in the rig description", block.Name)
		}
		if _, duplicate := seen[block.Name]; duplicate {
			return fmt.Errorf("block %q appears more than once", block.Name)
		}
		seen[block.Name] = struct{}{}
		if block.Path < 0 || block.Path > 1 || (!isDualDSP && block.Path != 0) {
			return fmt.Errorf("block %q has invalid path %d", block.Name, block.Path)
		}

		entry, found := helix.DB.FindByRealName(block.ModelName)
		if !found {
			entry, found = helix.DB.FindByID(block.ModelName)
		}
		if !found {
			return fmt.Errorf("block %q uses unknown model %q", block.Name, block.ModelName)
		}
		if !helix.IsCompatibleComponentModel(componentType, entry) {
			return fmt.Errorf("block %q uses model %q incompatible with component type %q", block.Name, block.ModelName, componentType)
		}
		pathDSP[block.Path] += helix.EffectiveDSPMono(entry)
		if err := validateParameterMap(entry, block.Params, false); err != nil {
			return fmt.Errorf("block %q: %w", block.Name, err)
		}
		blockModels[block.Name] = entry
	}
	for _, snapshot := range rig.Snapshots {
		for blockName, rawOverrides := range snapshot.Params {
			entry, mappable := blockModels[blockName]
			if !mappable {
				// Variax overrides are handled as global input settings, not block parameters.
				continue
			}
			overrides, ok := rawOverrides.(map[string]interface{})
			if !ok {
				return fmt.Errorf("snapshot %q block %q parameters must be an object", snapshot.Name, blockName)
			}
			if err := validateParameterMap(entry, overrides, true); err != nil {
				return fmt.Errorf("snapshot %q block %q: %w", snapshot.Name, blockName, err)
			}
		}
	}
	for path, cost := range pathDSP {
		if cost > helix.SafeDSPPerPath {
			return fmt.Errorf("path %d requires %.1f%% DSP, exceeding the safe %.1f%% budget", path, cost, helix.SafeDSPPerPath)
		}
	}

	return nil
}

var parameterAliases = map[string][]string{
	"gain":   {"Drive", "LeadGain", "Lead Drive", "ChVol", "Master"},
	"drive":  {"Gain", "LeadDrive", "Lead Gain", "Overdrive"},
	"volume": {"ChVol", "Master", "Level"},
	"vol":    {"ChVol", "Master", "Level"},
	"mids":   {"Middle", "Mid"},
}

func validateParameterMap(entry helix.CatalogEntry, values map[string]interface{}, snapshot bool) error {
	return validateParameterMapWithSafety(entry, values, snapshot, true)
}

func validateCatalogDefaultParameterMap(entry helix.CatalogEntry, values map[string]interface{}) error {
	return validateParameterMapWithSafety(entry, values, false, false)
}

func validateParameterMapWithSafety(entry helix.CatalogEntry, values map[string]interface{}, snapshot, enforceSafety bool) error {
	for requestedName, value := range values {
		name, err := resolveParameterName(entry, requestedName, snapshot)
		if err != nil {
			return err
		}
		if err := validateParameterValue(entry, name, value, enforceSafety); err != nil {
			return err
		}
	}
	return nil
}

func resolveParameterName(entry helix.CatalogEntry, requestedName string, snapshot bool) (string, error) {
	if strings.HasPrefix(requestedName, "@") {
		return "", fmt.Errorf("parameter %q is reserved", requestedName)
	}
	defaults, ok := entry.Data["Defaults"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("model %q has invalid defaults", entry.InternalName)
	}
	name := requestedName
	if _, exists := defaults[name]; !exists {
		found := false
		for actualName := range defaults {
			if strings.EqualFold(actualName, requestedName) {
				name = actualName
				found = true
				break
			}
		}
		if !found {
			for _, alias := range parameterAliases[strings.ToLower(requestedName)] {
				if _, exists := defaults[alias]; exists {
					name = alias
					found = true
					break
				}
			}
		}
		if !found {
			return "", fmt.Errorf("unsupported parameter %q", requestedName)
		}
	}
	if snapshot && !isSnapshotControllable(entry, name) {
		return "", fmt.Errorf("parameter %q cannot be controlled by snapshots", requestedName)
	}
	return name, nil
}

func validateParameterValue(entry helix.CatalogEntry, name string, value interface{}, enforceSafety bool) error {
	defaults := entry.Data["Defaults"].(map[string]interface{})
	expected := defaults[name]
	switch expected.(type) {
	case float64:
		actual, ok := value.(float64)
		if !ok {
			return fmt.Errorf("parameter %q must be a number", name)
		}
		if math.IsNaN(actual) || math.IsInf(actual, 0) {
			return fmt.Errorf("parameter %q must be finite", name)
		}
		min, max := parameterRange(entry, name, enforceSafety)
		if actual < min || actual > max {
			return fmt.Errorf("parameter %q value %.3f is outside the allowed range %.3f to %.3f", name, actual, min, max)
		}
	case bool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("parameter %q must be a boolean", name)
		}
	case string:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter %q must be a string", name)
		}
	default:
		return fmt.Errorf("parameter %q has unsupported catalog type %T", name, expected)
	}
	return nil
}

func isSnapshotControllable(entry helix.CatalogEntry, name string) bool {
	controllers, ok := entry.Data["Controller_Dict"].(map[string]interface{})
	if !ok {
		return false
	}
	_, ok = controllers[name]
	return ok
}

func parameterRange(entry helix.CatalogEntry, name string, enforceSafety bool) (float64, float64) {
	min, max := math.Inf(-1), math.Inf(1)
	if controllers, ok := entry.Data["Controller_Dict"].(map[string]interface{}); ok {
		if controller, ok := controllers[name].(map[string]interface{}); ok {
			if value, ok := controller["@min"].(float64); ok {
				min = value
			}
			if value, ok := controller["@max"].(float64); ok {
				max = value
			}
		}
	}

	lowerName := strings.ToLower(name)
	if !enforceSafety {
		return min, max
	}
	isReverb := strings.HasPrefix(entry.InternalName, "HD2_Reverb") || strings.HasPrefix(entry.InternalName, "VIC_Reverb")
	isDelay := strings.HasPrefix(entry.InternalName, "HD2_Delay") || strings.HasPrefix(entry.InternalName, "HD2_DL4") || strings.HasPrefix(entry.InternalName, "VIC_Delay")
	if (isReverb && (lowerName == "decay" || lowerName == "verbdecay")) || (isDelay && lowerName == "verbdecay") {
		max = math.Min(max, 0.7)
	}
	if isDelay && (lowerName == "feedback" || lowerName == "fdbk" || lowerName == "bk") {
		max = math.Min(max, 0.75)
	}
	return min, max
}
