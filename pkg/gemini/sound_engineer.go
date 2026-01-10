package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

// RigDescription is the output of the Designer Agent
type RigDescription struct {
	SuggestedName string         `json:"suggested_name"` // Concise name for the preset
	Explanation   string         `json:"explanation"`    // Textual description of the design
	Chain         []RigComponent `json:"chain"`
}

type RigComponent struct {
	Type        string `json:"type"`        // amp, cab, pedal, modulation, delay, reverb
	Name        string `json:"name"`        // Real world name, e.g. "Tube Screamer"
	Description string `json:"description"` // Brief motivation, e.g. "For mid boost"
	Settings    string `json:"settings"`    // Abstract settings, e.g. "High gain, low mids"
}

// ChatSoundEngineer creates or refines the abstract sound design based on discussion history
func (c *Client) ChatSoundEngineer(ctx context.Context, history []ChatMessage) (*RigDescription, error) {
	// Prompt engineering for Sound Engineer Agent
	sysPrompt := `You are a world-class Sound Engineer and guitar technician. 
	Your goal is to design or refine a guitar rig (signal chain) based on the user's description and the ongoing discussion.
	
	CONVERSATION LOGIC:
	- You are in a discussion stage. The user might ask for changes ("add more gain", "swap the amp for a Marshall").
	- Always provide a complete, updated RigDescription based on the current state of the discussion.
	- Your explanation should respond to the user's latest comments while keeping the context of the whole discussion.

	OUTPUT FORMAT:
	Return ONLY a JSON object with these top-level keys:
	1. "suggested_name": A VERY CONCISE name for the preset (MAX 16 characters). Based on the prompt. (e.g. "MAYER BLUES", "EVH BROWN").
	2. "explanation": A conversational, textual description of the sound design you created or modified. Explain WHY you made these recent changes.
	3. "chain": An array of components representing the ENTIRE signal chain.
	
	Each item in the "chain" should have:
	- "type": one of [pedal, amp, cab, modulation, delay, reverb]
	- "name": The SPECIFIC REAL-WORLD model name of the gear (e.g. "Ibanez Tube Screamer", "Fender Deluxe Reverb", "Vox AC30", "Klon Centaur"). DO NOT use Line 6 invented names.
	- "description": Why you chose this or how it fits the recent request.
	- "settings": A brief text description of how to dial it in (e.g. "Edge of breakup", "Long decay").

	Ensure the chain is logically ordered (Pedals -> Amp -> Cab -> Post-FX).
	If the user doesn't specify an amp, CHOOSE ONE that fits the style. ALWAYS include an Amp and a Cab.
	`

	model := c.client.GenerativeModel(c.ModelName)
	model.ResponseMIMEType = "application/json"

	// Construct the prompt parts including history
	var parts []genai.Part
	parts = append(parts, genai.Text(sysPrompt))

	for _, msg := range history {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		parts = append(parts, genai.Text(fmt.Sprintf("%s: %s", role, msg.Content)))
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("sound engineer agent failed: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Sound Engineer Agent")
	}

	var jsonText string
	if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		jsonText = string(txt)
	} else {
		return nil, fmt.Errorf("unexpected response format from Sound Engineer Agent")
	}

	var result RigDescription
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Sound Engineer JSON: %v. Raw: %s", err, jsonText)
	}

	return &result, nil
}
