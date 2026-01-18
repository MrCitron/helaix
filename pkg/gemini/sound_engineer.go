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
}

type RigComponent struct {
	Type        string `json:"type"`        // amp, cab, pedal, modulation, delay, reverb
	Name        string `json:"name"`        // Real world name, e.g. "Tube Screamer"
	Description string `json:"description"` // Brief motivation, e.g. "For mid boost"
	Settings    string `json:"settings"`    // Abstract settings, e.g. "High gain, low mids"
}

// ChatSoundEngineer creates or refines the abstract sound design based on discussion history
func (c *Client) ChatSoundEngineer(ctx context.Context, history []ChatMessage, hardwareModel string) (*RigDescription, error) {
	// Prompt engineering for Sound Engineer Agent
	sysPrompt := fmt.Sprintf(`You are a world-class Sound Engineer and guitar technician. 
	Your goal is to design or refine a guitar rig (signal chain) based on the user's description and the ongoing discussion.
	The user is using a **Line 6 Variax %s** hardware model.
	
	CONVERSATION LOGIC:
	- You are in a discussion stage. The user might ask for changes ("add more gain", "swap the amp for a Marshall").
	- Always provide a complete, updated RigDescription based on the current state of the discussion.
	- Your explanation should respond to the user's latest comments while keeping the context of the whole discussion.

	OUTPUT FORMAT:
	Return ONLY a JSON object with these top-level keys:
	1. "suggested_name": A VERY CONCISE name for the preset (MAX 16 characters). Based on the prompt. (e.g. "MAYER BLUES", "EVH BROWN").
	2. "explanation": A conversational, textual description of the sound design you created or modified. Explain WHY you made these recent changes.
	3. "guitar_model": A recommended **Real-World Guitar Model** name (e.g. "Stratocaster" for Strat-style, "Les Paul" for Gibson-style, "Telecaster" for Tele-style, "J-200" for acoustic). Use iconic names.
	4. "tuning": A specific tuning required for the style, requested by the user, or contextually appropriate for a specific song (e.g. "Standard", "Drop D", "Eb Standard", "D Standard", "Drop C"). If no specific tuning is needed, leave this EMPTY.
	5. "chain": An array of components representing the ENTIRE signal chain.
	
	Each item in the "chain" should have:
	- "type": one of [pedal, amp, cab, modulation, delay, reverb]
	- "name": The SPECIFIC REAL-WORLD model name of the gear (e.g. "Ibanez Tube Screamer", "Fender Deluxe Reverb", "Vox AC30", "Klon Centaur"). DO NOT use Line 6 invented names.
	- "description": Why you chose this or how it fits the recent request.
	- "settings": A brief text description of how to dial it in (e.g. "Edge of breakup", "Long decay").

	GUITAR & TUNING LOGIC:
	- If the user asks for a specific artist sound, suggest the guitar model and tuning they are famous for (e.g. Gilmour -> Stratocaster, Page -> Les Paul).
	- For heavy rock/metal (e.g. Rage Against The Machine), favor "Les Paul" or "Humbucker" styles unless single-coils are essential to the specific song (e.g. "Killing in the Name" uses a Telecaster).
	- Suggest a tuning if the context or requested song/style demands it (e.g. "Rage Against The Machine" or "Drop D style" -> "Drop D").
	- If the user specifies a tuning (e.g. "Baritone tuning"), provide the appropriate name for it.
	- If no special tuning is required for the style/song, leave "tuning" EMPTY.

	Ensure the chain is logically ordered (Pedals -> Amp -> Cab -> Post-FX).
	If the user doesn't specify an amp, CHOOSE ONE that fits the style. ALWAYS include an Amp and a Cab.
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
