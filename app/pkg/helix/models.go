package helix

// Note: This is a best-effort reverse-engineered schema.
// Real .hlx files are more complex.

// Preset is now defined in preset.go as map[string]interface{}

// Block represents a simplified view of an effect block for AI generation purposes
type AIBlock struct {
	Model  string                 `json:"model"`
	Path   int                    `json:"path"` // 0 or 1
	ID     string                 `json:"id"`   // e.g. "block0"
	Params map[string]interface{} `json:"params"`
}

// Helper to construct a base preset
