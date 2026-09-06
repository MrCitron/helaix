package gemini

import (
	"HelAIx/pkg/helix"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

type toneFixture struct {
	Name          string              `json:"name"`
	Rig           RigDescription      `json:"rig"`
	ExpectedHints map[string][]string `json:"expected_hints"`
}

type fixtureCandidateGroup struct {
	component  RigComponent
	candidates []helix.CatalogEntry
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
		"clean": {}, "crunch": {}, "high-gain": {}, "ambient": {}, "metal": {}, "acoustic": {},
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
			search(index+1, score+rank, usedDSP+helix.EffectiveDSPMono(candidate))
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
