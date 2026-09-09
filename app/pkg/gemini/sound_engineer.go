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
	Instrument    string         `json:"instrument"`     // Resolved instrument: Guitar or Bass
	UseVariax     bool           `json:"use_variax"`     // Resolved Variax decision
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
func (c *Client) ChatSoundEngineer(ctx context.Context, history []ChatMessage, hardwareModel string, defaultInstrument string, variaxEnabled bool) (*RigDescription, error) {
	// Prompt engineering for Sound Engineer Agent
	sysPrompt := fmt.Sprintf(`You are a world-class Sound Engineer and guitar/bass technician.
	Your goal is to design or refine a signal chain based on the user's description and the ongoing discussion.
	The user's default instrument context is: **%s**.
	The user is using a **Line 6 Variax %s** hardware model (if enabled).
	Automatic Variax control is configured: **%t**.
	
	CONVERSATION LOGIC:
	- You are in a refinement loop. The user might ask for changes ("add more gain", "swap the amp").
	- Always provide a complete, updated RigDescription based on the current state.
	- STABILITY RULE: You MUST keep the exact "name" and "type" of every component that the user did NOT ask to change. Never rename gear unless specifically asked (e.g. don't change "Tube Screamer" to "Ibanez TS9" midway through).
	- Your explanation should respond to the user's latest comments while keeping the context of the whole discussion.

	OUTPUT FORMAT:
	Return ONLY a JSON object with these top-level keys:
	1. "suggested_name": A VERY CONCISE name for the preset (MAX 16 characters). Based on the prompt. (e.g. "MAYER BLUES", "EVH BROWN").
	2. "explanation": A conversational, textual description of the sound design you created or modified. Explain WHY you made these recent changes.
	3. "instrument": The resolved instrument, exactly "Guitar" or "Bass".
	4. "use_variax": true only when the resolved instrument is Guitar and Automatic Variax control is configured true.
	5. "guitar_model": A recommended real-world instrument model name. Use a bass model for Bass and a guitar model for Guitar.
	6. "tuning": A specific tuning required (e.g. "Standard"). Global default for the preset.
	7. "chain": An array of components representing the ENTIRE signal chain.
	8. "snapshots": (Conditional) An array of 1 to 4 snapshot objects if a song/artist is requested or explicitly asked for.
	
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
	- **INSTRUMENT RESOLUTION**: If the user's prompt explicitly mentions Guitar or Bass, that overrides the default instrument. Otherwise use the default instrument from settings. Do not infer a different instrument from genre, artist, or tone descriptions alone.
	- If the context is Bass, explicitly prioritize bass amp models (e.g., SVT, Ampeg, GK) and bass cabs.
	- **VARIAX DECISION**: Set "use_variax" to true only when the resolved instrument is Guitar and Automatic Variax control is configured true. If the resolved instrument is Bass, set "use_variax" to false regardless of the setting.
	- **VARIAX BASS CONSTRAINT**: A Variax guitar setting must never be used to simulate a bass. For a resolved Bass design, do not add a Variax component and use standard bass modeling for a physical bass instrument.
	- **REAL-WORLD NAMES ONLY**: The "guitar_model" fields must use iconic, real-world instrument names (e.g. "Fender Stratocaster", "Fender Precision Bass").
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
	- Add a "Line6 Variax" component to the beginning of the "chain" array only when "use_variax" is true. If it is false, DO NOT add a Variax component.

	Ensure the chain is logically ordered (Pedals -> Amp -> Cab -> Post-FX).
	ALWAYS include an Amp and a Cab.
	`, defaultInstrument, hardwareModel, variaxEnabled)

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
	if result.Instrument != "Guitar" && result.Instrument != "Bass" {
		result.Instrument = defaultInstrument
	}
	result.UseVariax = resolveVariaxDecision(result.Instrument, variaxEnabled)

	return &result, nil
}

// resolveVariaxDecision allows Variax only for guitars when automatic control is enabled.
func resolveVariaxDecision(instrument string, enabled bool) bool {
	return instrument == "Guitar" && enabled
}
