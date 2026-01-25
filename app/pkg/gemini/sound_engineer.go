package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

// RigDescription is the output of the Designer Agent
type RigDescription struct {
	SuggestedName string         `json:"suggested_name"` // Concise name for the preset
	Explanation   string         `json:"explanation"`    // Textual description of the design
	GuitarModel   string         `json:"guitar_model"`   // Variax model suggestion (Lester, Spank, T-Model, Acoustic, etc)
	Tuning        string         `json:"tuning"`         // Tuning suggestion (Standard, Drop D, Half Step Down, etc)
	Chain         []RigComponent `json:"chain"`
	Snapshots     []Snapshot     `json:"snapshots,omitempty"`
}

type RigComponent struct {
	Type        string `json:"type"`        // amp, cab, pedal, modulation, delay, reverb
	Name        string `json:"name"`        // Real world name, e.g. "Tube Screamer"
	Description string `json:"description"` // Brief motivation, e.g. "For mid boost"
	Settings    string `json:"settings"`    // Abstract settings, e.g. "High gain, low mids"
}

type Snapshot struct {
	Name         string                 `json:"name"`
	ActiveBlocks []string               `json:"active_blocks"` // Names of blocks that are enabled in this snapshot
	GuitarModel  string                 `json:"guitar_model,omitempty"`
	Tuning       string                 `json:"tuning,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"` // BlockName -> { "Param": Value }
}

// ChatSoundEngineer creates or refines the abstract sound design based on discussion history
func (c *Client) ChatSoundEngineer(ctx context.Context, history []ChatMessage, hardwareModel string) (*RigDescription, error) {
	// Prompt engineering for Sound Engineer Agent
	sysPrompt := fmt.Sprintf(`You are a world-class Sound Engineer and guitar technician. 
	Your goal is to design or refine a guitar rig (signal chain) based on the user's description and the ongoing discussion.
	The user is using a **Line 6 Variax %s** hardware model.
	
	CONVERSATION LOGIC:
	- You are in a refinement loop. The user might ask for changes ("add more gain", "swap the amp").
	- Always provide a complete, updated RigDescription based on the current state.
	- STABILITY RULE: You MUST keep the exact "name" and "type" of every component that the user did NOT ask to change. Never rename gear unless specifically asked (e.g. don't change "Tube Screamer" to "Ibanez TS9" midway through).
	- Your explanation should respond to the user's latest comments while keeping the context of the whole discussion.

	OUTPUT FORMAT:
	Return ONLY a JSON object with these top-level keys:
	1. "suggested_name": A VERY CONCISE name for the preset (MAX 16 characters). Based on the prompt. (e.g. "MAYER BLUES", "EVH BROWN").
	2. "explanation": A conversational, textual description of the sound design you created or modified. Explain WHY you made these recent changes.
	3. "guitar_model": A recommended **Real-World Guitar Model** name (e.g. "Stratocaster"). Global default for the preset.
	4. "tuning": A specific tuning required (e.g. "Standard"). Global default for the preset.
	5. "chain": An array of components representing the ENTIRE signal chain.
	6. "snapshots": (Conditional) An array of 1 to 4 snapshot objects if a song/artist is requested or explicitly asked for.
	
	Each "chain" item should have:
	- "type": one of [pedal, amp, cab, modulation, delay, reverb, variax]
	- "name": The SPECIFIC REAL-WORLD model name of the gear (e.g. "Ibanez Tube Screamer"). For variax, use "Line6 Variax".
	- "description": Why you chose this or how it fits.
	- "settings": A brief text description of how to dial it in (e.g. "Lester model, Standard tuning").

	Each "snapshot" item should have:
	- "name": Concise part name (e.g. "Intro", "Chorus", "Solo", "Clean", "Lead").
	- "active_blocks": An array of "name" strings from the "chain" that should be ENABLED in this snapshot. Others will be DISABLED.
	- "guitar_model": (Optional) Override the global guitar model for this snapshot.
	- "tuning": (Optional) Override the global tuning for this snapshot.
	- "params": (Optional) A map where keys are block names and values are objects of parameter overrides (e.g. {"Marshall Plexi": {"Drive": 0.8, "Master": 1.0}}). Only specify what MUST change compared to the baseline.

	SNAPSHOT LOGIC:
	- **Condition**: Generate snapshots ONLY if the user asks for a specific song/artist or explicitly requests them.
	- **Consistency**: EVERY block in the "chain" must be enabled in AT LEAST one snapshot. Do not include blocks that are never used.
	- **Exclusion**: For "Clean" or "Clean/Verse" snapshots, you MUST disable high-gain Distortion/Overdrive blocks, but you should usually keep Modulation (Chorus/Flanger), Reverb, and Delay ENABLED if they contribute to the clean texture.
	- **Accuracy**: The strings in "active_blocks" MUST match the "name" field in the "chain" EXACTLY.
	- **Limit**: Strictly maximum 4 snapshots.

	GUITAR & VARIAX LOGIC:
	- **REAL-WORLD NAMES ONLY**: The "guitar_model" fields must use iconic, real-world guitar names (e.g. "Fender Stratocaster", "Gibson Les Paul Standard", "Fender Jaguar"). 
	- **FORBIDDEN**: Never use technical Variax bank names like "Spank", "Lester", or "T-Model" in these fields. 
	- **SNAP-LOCK REQUIREMENT**: YOU MUST populate the "guitar_model" field for EVERY snapshot.
	- **VARIANT SPECIFICATION**: To select a specific variant (1-5), append the pickup position in parentheses: "Fender Stratocaster (Pickup Pos 2)". 
	- **HARDWARE MAPPING REFERENCE**:
	  - Fender Stratocaster -> Bank: Spank
	  - Fender Telecaster / Jaguar -> Bank: T-Model
	  - Gibson Les Paul -> Bank: Lester
	  - Gibson LP Special / Firebird -> Bank: Special
	  - Gretsch / Duo Jet -> Bank: R-Billy
	  - Rickenbacker -> Bank: Chime
	  - Hollowbody / ES-335 -> Bank: Semi
	  - Archtop / Jazzbox -> Bank: Jazzbox
	  - Acoustic Martin/Gibson -> Bank: Acoustic
	  - Sitar / Banjo / Dobro -> Bank: Reso
	- **JSON EXAMPLE**:
	  {
	    "guitar_model": "Fender Stratocaster (Pickup Pos 1)", 
	    "snapshots": [
	      { "name": "Intro", "guitar_model": "Fender Jaguar (Pickup Pos 1)", "active_blocks": [...] },
	      { "name": "Chorus", "guitar_model": "Gibson Les Paul (Pickup Pos 5)", "active_blocks": [...] }
	    ]
	  }
	- Additionally, YOU SHOULD add a "Line6 Variax" component to the beginning of the "chain" array. Store the variant in its "settings" field (e.g. "Fender Jaguar").

	Ensure the chain is logically ordered (Pedals -> Amp -> Cab -> Post-FX).
	ALWAYS include an Amp and a Cab.
	`, hardwareModel)

	// Construct the conversation history with system prompt
	var contents []*genai.Content

	// Add system instruction as first user message
	contents = append(contents, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: sysPrompt}},
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
		return nil, fmt.Errorf("sound engineer agent failed: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Sound Engineer Agent")
	}

	jsonText := resp.Candidates[0].Content.Parts[0].Text

	var result RigDescription
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Sound Engineer JSON: %v. Raw: %s", err, jsonText)
	}

	return &result, nil
}
