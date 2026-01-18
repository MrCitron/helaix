package helix

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed data/catalog.json
var catalogJSON []byte

// Global instance
var DB CatalogDB

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
