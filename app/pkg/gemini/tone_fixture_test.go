package gemini

import (
	"HelAIx/pkg/helix"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

type toneFixture struct {
	Name           string              `json:"name"`
	Rig            RigDescription      `json:"rig"`
	ExpectedHints  map[string][]string `json:"expected_hints"`
	ExpectedRecipe *recipeExpectation  `json:"expected_recipe"`
}

type recipeExpectation struct {
	Models map[string]string                 `json:"models"`
	Params map[string]map[string]interface{} `json:"params"`
}

type fixtureCandidateGroup struct {
	component  RigComponent
	candidates []helix.CatalogEntry
}

func TestCuratedToneFixtures(t *testing.T) {
	for _, fixture := range loadToneFixtures(t) {
		if fixture.ExpectedRecipe == nil {
			continue
		}
		t.Run(fixture.Name, func(t *testing.T) {
			plan, err := recommendedPlanForRig(fixture.Rig, false)
			if err != nil {
				t.Fatalf("recommendedPlanForRig() error = %v", err)
			}
			if plan.dsp[0] > helix.SafeDSPPerPath {
				t.Fatalf("recipe uses %.1f%% DSP, exceeding %.1f%%", plan.dsp[0], helix.SafeDSPPerPath)
			}

			response := builderResponse{}
			for _, block := range plan.blocks {
				response.Blocks = append(response.Blocks, builderBlock{Name: block.name, ModelName: block.modelName, Path: block.path, Params: block.params})
			}
			assertRecipeExpectation(t, response, *fixture.ExpectedRecipe)
			if err := validateBuilderResponse(response, &fixture.Rig, false); err != nil {
				t.Fatalf("recipe response is invalid: %v", err)
			}

			modelOnly := builderResponse{}
			for _, block := range response.Blocks {
				modelOnly.Blocks = append(modelOnly.Blocks, builderBlock{Name: block.Name, ModelName: block.ModelName, Path: block.Path})
			}
			jsonText, err := json.Marshal(modelOnly)
			if err != nil {
				t.Fatalf("marshal recipe response: %v", err)
			}
			preset, err := BuildPresetFromJSON(string(jsonText), &fixture.Rig, fixture.Name, "Helix Stomp", 0, false, "Variax Standard")
			if err != nil {
				t.Fatalf("export recipe preset: %v", err)
			}
			if err := helix.ValidatePreset(*preset); err != nil {
				t.Fatalf("exported recipe preset is invalid: %v", err)
			}
			if _, err := json.Marshal(preset); err != nil {
				t.Fatalf("serialize recipe preset: %v", err)
			}
			assertExportedRecipeExpectation(t, exportedBlocks(t, preset), *fixture.ExpectedRecipe)
		})
	}
}

func TestToneFixtures(t *testing.T) {
	fixtures := loadToneFixtures(t)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			if err := validateRigDescription(&fixture.Rig); err != nil {
				t.Fatalf("fixture rig is invalid: %v", err)
			}

			response, totalDSP := fixtureBuilderResponse(t, fixture)
			if totalDSP > helix.SafeDSPPerPath {
				t.Fatalf("fixture selected %.1f%% DSP, exceeding %.1f%%", totalDSP, helix.SafeDSPPerPath)
			}
			if err := validateBuilderResponse(response, &fixture.Rig, false); err != nil {
				t.Fatalf("fixture builder response is invalid: %v", err)
			}
		})
	}
}

func loadToneFixtures(t *testing.T) []toneFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/tone_fixtures.json")
	if err != nil {
		t.Fatalf("read tone fixtures: %v", err)
	}
	var fixtures []toneFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("parse tone fixtures: %v", err)
	}
	if len(fixtures) != 6 {
		t.Fatalf("fixture count = %d, want 6", len(fixtures))
	}
	expectedNames := map[string]struct{}{
		"clean": {}, "crunch": {}, "high-gain": {}, "ambient": {}, "tight-metal": {}, "acoustic": {},
	}
	for _, fixture := range fixtures {
		if _, exists := expectedNames[fixture.Name]; !exists {
			t.Fatalf("unexpected or duplicate fixture %q", fixture.Name)
		}
		delete(expectedNames, fixture.Name)
	}
	if len(expectedNames) != 0 {
		t.Fatalf("missing fixtures: %v", expectedNames)
	}
	return fixtures
}

func fixtureBuilderResponse(t *testing.T, fixture toneFixture) (builderResponse, float64) {
	t.Helper()
	groups := make([]fixtureCandidateGroup, 0, len(fixture.Rig.Chain))
	for _, component := range fixture.Rig.Chain {
		if component.Type == "variax" {
			continue
		}
		intent := strings.Join([]string{component.Name, component.Description, component.Settings}, " ")
		candidates := helix.RankedCandidatesForComponent(component.Type, intent, helix.PromptCandidateLimit)
		if len(candidates) == 0 || len(candidates) > helix.PromptCandidateLimit {
			t.Fatalf("%s candidates = %d, want 1..%d", component.Name, len(candidates), helix.PromptCandidateLimit)
		}
		for _, candidate := range candidates {
			if !helix.IsCompatibleComponentModel(component.Type, candidate) {
				t.Fatalf("%s candidate %q is incompatible with %s", component.Name, candidate.InternalName, component.Type)
			}
		}
		if hints := fixture.ExpectedHints[component.Name]; len(hints) > 0 && !topCandidatesMatch(candidates, hints) {
			t.Fatalf("%s top candidates %q do not match any tonal hint %v", component.Name, candidateNames(candidates), hints)
		}
		groups = append(groups, fixtureCandidateGroup{component: component, candidates: candidates})
	}

	selected, totalDSP, ok := bestDSPFixtureCombination(groups)
	if !ok {
		t.Fatalf("fixture has no ranked candidate combination within the %.1f%% DSP budget", helix.SafeDSPPerPath)
	}
	response := builderResponse{}
	for index, candidate := range selected {
		component := groups[index].component
		modelName := candidate.Name
		if modelName == "" {
			modelName = candidate.InternalName
		}
		response.Blocks = append(response.Blocks, builderBlock{Name: component.Name, ModelName: modelName, Path: 0})
	}
	return response, totalDSP
}

func topCandidatesMatch(candidates []helix.CatalogEntry, hints []string) bool {
	limit := min(3, len(candidates))
	for _, candidate := range candidates[:limit] {
		document := strings.ToLower(strings.Join([]string{candidate.Name, candidate.BasedOn, candidate.InternalName}, " "))
		for _, hint := range hints {
			if strings.Contains(document, hint) {
				return true
			}
		}
	}
	return false
}

func bestDSPFixtureCombination(groups []fixtureCandidateGroup) ([]helix.CatalogEntry, float64, bool) {
	bestScore := int(^uint(0) >> 1)
	bestDSP := 0.0
	var best []helix.CatalogEntry
	selected := make([]helix.CatalogEntry, len(groups))
	var search func(index, score int, usedDSP float64)
	search = func(index, score int, usedDSP float64) {
		if usedDSP > helix.SafeDSPPerPath || score > bestScore {
			return
		}
		if index == len(groups) {
			if score < bestScore || (score == bestScore && usedDSP < bestDSP) {
				bestScore = score
				bestDSP = usedDSP
				best = append([]helix.CatalogEntry(nil), selected...)
			}
			return
		}
		for rank, candidate := range groups[index].candidates {
			selected[index] = candidate
			search(index+1, score+rank, usedDSP+helix.EffectiveDSP(candidate))
		}
	}
	search(0, 0, 0)
	return best, bestDSP, len(best) > 0
}

func candidateNames(candidates []helix.CatalogEntry) string {
	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		name := candidate.Name
		if name == "" {
			name = candidate.InternalName
		}
		names = append(names, name)
	}
	return fmt.Sprintf("%v", names)
}

func assertRecipeExpectation(t *testing.T, response builderResponse, expected recipeExpectation) {
	t.Helper()
	blocks := make(map[string]builderBlock, len(response.Blocks))
	for _, block := range response.Blocks {
		blocks[block.Name] = block
	}
	for name, model := range expected.Models {
		block, ok := blocks[name]
		if !ok || block.ModelName != model {
			t.Fatalf("recipe block %q model = %q, want %q", name, block.ModelName, model)
		}
		for parameter, value := range expected.Params[name] {
			if !reflect.DeepEqual(block.Params[parameter], value) {
				t.Fatalf("recipe block %q parameter %q = %v, want %v", name, parameter, block.Params[parameter], value)
			}
		}
	}
}

func exportedBlocks(t *testing.T, preset *helix.Preset) map[string]map[string]interface{} {
	t.Helper()
	data := (*preset)["data"].(map[string]interface{})
	tone := data["tone"].(map[string]interface{})
	blocks := make(map[string]map[string]interface{})
	for _, path := range []string{"dsp0", "dsp1"} {
		for _, rawBlock := range tone[path].(map[string]interface{}) {
			block, ok := rawBlock.(map[string]interface{})
			if !ok {
				continue
			}
			if name, ok := block["@name"].(string); ok {
				blocks[name] = block
			}
		}
	}
	return blocks
}

func assertExportedRecipeExpectation(t *testing.T, blocks map[string]map[string]interface{}, expected recipeExpectation) {
	t.Helper()
	for name, model := range expected.Models {
		block, ok := blocks[name]
		if !ok {
			t.Fatalf("exported recipe is missing block %q", name)
		}
		entry, found := helix.DB.FindByRealName(model)
		if !found || block["@model"] != entry.InternalName {
			t.Fatalf("exported recipe block %q model = %v, want %q", name, block["@model"], entry.InternalName)
		}
		for parameter, value := range expected.Params[name] {
			if !reflect.DeepEqual(block[parameter], value) {
				t.Fatalf("exported recipe block %q parameter %q = %v, want %v", name, parameter, block[parameter], value)
			}
		}
	}
}
