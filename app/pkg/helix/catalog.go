package helix

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"unicode"
)

//go:embed data/catalog.json
var catalogJSON []byte

// Global instance
var DB CatalogDB

const (
	DefaultDSPMono       = 3.0
	SafeDSPPerPath       = 65.0
	PromptCandidateLimit = 8
)

type toneProfile struct {
	signals    []string
	modelHints []string
}

var toneProfiles = []toneProfile{
	{signals: []string{"clean", "limpio", "cristalino", "crystal", "jazz"}, modelHints: []string{"clean", "deluxe", "twin", "princess", "rivet", "jazz", "aristocrat", "litigator", "fullerton"}},
	{signals: []string{"crunch", "crujiente", "blues", "rock", "clasico", "classic"}, modelHints: []string{"crunch", "plexi", "brit", "matchstick", "tweed", "essex", "a30", "rocker"}},
	{signals: []string{"high gain", "alta ganancia", "metal", "heavy", "djent", "brown", "moderno", "modern"}, modelHints: []string{"lead", "rectifire", "vitriol", "panama", "ubersonic", "badonk", "fatality", "placater", "revv", "meteor", "solo", "cartographer"}},
	{signals: []string{"ambient", "ambiental", "ethereal", "etereo", "shimmer", "espacioso"}, modelHints: []string{"shimmer", "glitz", "ganymede", "plateaux", "searchlights", "particle", "cave", "cosmos", "heliosphere", "adriatic", "tape"}},
	{signals: []string{"vintage", "retro", "analog", "analogo", "analogico"}, modelHints: []string{"vintage", "analog", "bucket", "tape", "spring", "tweed", "plexi"}},
	{signals: []string{"digital", "moderno", "modern", "preciso", "precision"}, modelHints: []string{"digital", "cosmos", "transistor", "glitch", "horizon", "precision"}},
}

var toneTextReplacer = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n", "-", " ")

var toneStopWords = map[string]struct{}{
	"amp": {}, "and": {}, "cab": {}, "con": {}, "del": {}, "effect": {}, "for": {}, "from": {}, "gain": {}, "low": {}, "los": {}, "mas": {}, "more": {}, "pedal": {}, "para": {}, "por": {}, "the": {}, "una": {}, "use": {}, "with": {},
}

// CatalogEntry matches the generated JSON structure
type CatalogEntry struct {
	InternalName string                 `json:"InternalName"`
	Name         string                 `json:"Name"`
	BasedOn      string                 `json:"BasedOn"`
	DSPMono      float64                `json:"DSP_Mono,omitempty"`
	DSPStereo    float64                `json:"DSP_Stereo,omitempty"`
	Data         map[string]interface{} `json:"Data"` // Full block data
}

type CatalogDB struct {
	Entries []CatalogEntry
	// Indexes for fast lookup
	byInternalName map[string]CatalogEntry
	byName         map[string]CatalogEntry // Display Name -> Entry
	once           sync.Once
}

func (db *CatalogDB) EnsureLoaded() {
	db.once.Do(func() {
		if err := json.Unmarshal(catalogJSON, &db.Entries); err != nil {
			panic("Failed to load embedded catalog.json: " + err.Error())
		}
		db.byInternalName = make(map[string]CatalogEntry)
		db.byName = make(map[string]CatalogEntry)

		for _, e := range db.Entries {
			db.byInternalName[e.InternalName] = e
			if e.Name != "" {
				db.byName[strings.ToLower(e.Name)] = e
			}
		}
	})
}

// FindByRealName attempts to find a model by its display name (case insensitive)
func (db *CatalogDB) FindByRealName(name string) (CatalogEntry, bool) {
	db.EnsureLoaded()
	entry, ok := db.byName[strings.ToLower(name)]
	return entry, ok
}

// FindByID finds by Internal Name
func (db *CatalogDB) FindByID(id string) (CatalogEntry, bool) {
	db.EnsureLoaded()
	entry, ok := db.byInternalName[id]
	return entry, ok
}

// GetAllModels returns a list of "Real Name (Based On)" for the AI prompt
func (db *CatalogDB) GetAllModels() []string {
	db.EnsureLoaded()
	var list []string
	for _, e := range db.Entries {
		if e.Name != "" {
			list = append(list, e.Name)
		} else {
			list = append(list, e.InternalName)
		}
	}
	return list
}

// IsValidModel checks if id exists
func IsValidModel(id string) bool {
	DB.EnsureLoaded()
	_, ok := DB.byInternalName[id]
	return ok
}

// EffectiveDSPMono returns the catalog cost or a conservative fallback for unmeasured models.
func EffectiveDSPMono(entry CatalogEntry) float64 {
	if entry.DSPMono > 0 {
		return entry.DSPMono
	}
	return DefaultDSPMono
}

// CandidatesForComponent returns the Helix models that can implement one rig component type.
// The catalog's internal model families are the authoritative category metadata.
func CandidatesForComponent(componentType string) []CatalogEntry {
	DB.EnsureLoaded()
	candidates := make([]CatalogEntry, 0)
	for _, entry := range DB.Entries {
		if IsCompatibleComponentModel(componentType, entry) {
			candidates = append(candidates, entry)
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return modelLabel(candidates[i]) < modelLabel(candidates[j])
	})
	return candidates
}

// RankedCandidatesForComponent returns the best type-compatible models for a tone request.
func RankedCandidatesForComponent(componentType, toneIntent string, limit int) []CatalogEntry {
	return rankCandidates(CandidatesForComponent(componentType), toneIntent, limit)
}

func rankCandidates(candidates []CatalogEntry, toneIntent string, limit int) []CatalogEntry {
	if limit <= 0 || limit > len(candidates) {
		limit = len(candidates)
	}
	type scoredCandidate struct {
		entry CatalogEntry
		score int
	}
	query := normalizeToneText(toneIntent)
	terms := toneTerms(query)
	scored := make([]scoredCandidate, len(candidates))
	for i, candidate := range candidates {
		document := normalizeToneText(strings.Join([]string{candidate.Name, candidate.BasedOn, candidate.InternalName}, " "))
		score := 0
		for _, term := range terms {
			if strings.Contains(document, term) {
				score += 2
			}
		}
		for _, profile := range toneProfiles {
			if !containsAny(query, profile.signals) {
				continue
			}
			for _, hint := range profile.modelHints {
				if strings.Contains(document, hint) {
					score += 6
				}
			}
		}
		scored[i] = scoredCandidate{entry: candidate, score: score}
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return modelLabel(scored[i].entry) < modelLabel(scored[j].entry)
		}
		return scored[i].score > scored[j].score
	})
	ranked := make([]CatalogEntry, limit)
	for i := range ranked {
		ranked[i] = scored[i].entry
	}
	return ranked
}

func normalizeToneText(text string) string {
	return toneTextReplacer.Replace(strings.ToLower(text))
}

func toneTerms(text string) []string {
	seen := make(map[string]struct{})
	terms := make([]string, 0)
	for _, term := range strings.FieldsFunc(text, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) }) {
		if len(term) < 3 {
			continue
		}
		if _, ignored := toneStopWords[term]; ignored {
			continue
		}
		if _, exists := seen[term]; !exists {
			seen[term] = struct{}{}
			terms = append(terms, term)
		}
	}
	return terms
}

func containsAny(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

// IsCompatibleComponentModel reports whether a catalog model belongs to a rig component type.
func IsCompatibleComponentModel(componentType string, entry CatalogEntry) bool {
	id := entry.InternalName
	switch strings.ToLower(strings.TrimSpace(componentType)) {
	case "amp":
		return strings.HasPrefix(id, "HD2_Amp")
	case "cab":
		return strings.HasPrefix(id, "HD2_CabMicIr") || strings.HasPrefix(id, "VIC_Cab")
	case "delay":
		return strings.HasPrefix(id, "HD2_Delay") || strings.HasPrefix(id, "HD2_DL4")
	case "reverb":
		return strings.HasPrefix(id, "HD2_Reverb") || strings.HasPrefix(id, "HD2_DynPlate") || strings.HasPrefix(id, "VIC_Reverb")
	case "modulation":
		return hasModelPrefix(id,
			"HD2_Chorus", "HD2_Flanger", "HD2_FlexoVibe", "HD2_M13", "HD2_MM4",
			"HD2_Phaser", "HD2_Rotary", "HD2_Tremolo", "HD2_Vibrato",
		)
	case "pedal":
		return hasModelPrefix(id,
			"HD2_Dist", "HD2_Compressor", "HD2_DM4", "HD2_EQ", "HD2_FeedbackSim",
			"HD2_Filter", "HD2_FM4", "HD2_Gate", "HD2_Pitch", "HD2_RetroReel",
			"HD2_RingModulator", "HD2_VolPan", "HD2_Wah",
		)
	default:
		return false
	}
}

func hasModelPrefix(id string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

func modelLabel(entry CatalogEntry) string {
	if entry.Name != "" {
		return entry.Name
	}
	return entry.InternalName
}
