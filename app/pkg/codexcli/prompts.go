package codexcli

import (
	"HelAIx/pkg/gemini"
	"HelAIx/pkg/helix"
	"encoding/json"
	"fmt"
	"strings"
)

func soundPrompt(history []gemini.ChatMessage, hardwareModel string) string {
	return fmt.Sprintf(`You are HelAIx's guitar sound engineer. Design or refine a complete Line 6 Helix Floor rig for a Line 6 Variax %s.
Return only JSON that conforms to the supplied schema. suggested_name must contain 1 to 16 characters. Keep unchanged component names and types unless the user asks to change them. The chain must be ordered pedal -> amp -> cab -> post-FX, contain exactly one amp and one cab, and use 1 to 4 snapshots only when requested; return an empty snapshots array otherwise. Every snapshot must name only exact chain block names. Use real-world guitar names, not Variax bank names. Include Variax as a descriptive input component only; it is not an effect block.

Conversation history:
%s`, hardwareModel, historyJSON(history))
}

func presetPrompt(rig gemini.RigDescription, history []gemini.ChatMessage, hardware string) (string, error) {
	helix.DB.EnsureLoaded()
	var catalog strings.Builder
	for _, entry := range helix.DB.Entries {
		cost := entry.DSPMono
		if cost == 0 {
			cost = 3
		}
		catalog.WriteString(fmt.Sprintf("- %s (Based on: %s) [DSP: %.1f%%]\n", entry.Name, entry.BasedOn, cost))
	}
	rigJSON, err := json.Marshal(rig)
	if err != nil {
		return "", err
	}
	paths := "Use path 0 only."
	if strings.Contains(hardware, "Floor") || strings.Contains(hardware, "LT") || strings.Contains(hardware, "Rack") {
		paths = "Helix Floor has two DSP paths: use 0 or 1, keeping each near 60-65%% DSP."
	}
	return fmt.Sprintf(`You are HelAIx's Line 6 Helix preset engineer. Map every non-Variax component in this rig to an exact model_name from the catalog below. Return only JSON conforming to the supplied schema.
Hardware: %s. %s Never invent a model identifier. Keep stable choices from conversation unless the user requested a change. Variax is a global input and MUST NOT appear in blocks. Reverb Decay/VerbDecay must not exceed 0.7.

Rig: %s

Conversation history:
%s

Available Helix models:
%s`, hardware, paths, rigJSON, historyJSON(history), catalog.String()), nil
}

func historyJSON(history []gemini.ChatMessage) string {
	data, err := json.Marshal(history)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func rigSchema() []byte {
	return []byte(strings.ReplaceAll(`{"type":"object","additionalProperties":false,"required":["suggested_name","explanation","guitar_model","tuning","chain","snapshots"],"properties":{"suggested_name":{"type":"string"},"explanation":{"type":"string"},"guitar_model":{"type":"string"},"tuning":{"type":"string"},"chain":{"type":"array","items":{"type":"object","additionalProperties":false,"required":["type","name","description","settings"],"properties":{"type":{"type":"string","enum":["pedal","amp","cab","modulation","delay","reverb","variax"]},"name":{"type":"string"},"description":{"type":"string"},"settings":{"type":"string"}}}},"snapshots":{"type":"array","items":{"type":"object","additionalProperties":false,"required":["name","active_blocks","guitar_model","tuning"],"properties":{"name":{"type":"string"},"active_blocks":{"type":"array","items":{"type":"string"}},"guitar_model":{"type":"string"},"tuning":{"type":"string"}}}}}}`, `\`, ""))
}

func blocksSchema() []byte {
	return []byte(strings.ReplaceAll(`{"type":"object","additionalProperties":false,"required":["blocks"],"properties":{"blocks":{"type":"array","items":{"type":"object","additionalProperties":false,"required":["name","model_name","path"],"properties":{"name":{"type":"string"},"model_name":{"type":"string"},"path":{"type":"integer","enum":[0,1]}}}}}}`, `\`, ""))
}

func connectionSchema() []byte {
	return []byte(strings.ReplaceAll(`{"type":"object","additionalProperties":false,"required":["ok"],"properties":{"ok":{"type":"boolean"}}}`, `\`, ""))
}
