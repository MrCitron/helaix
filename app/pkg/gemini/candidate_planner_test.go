package gemini

import (
	"HelAIx/pkg/helix"
	"errors"
	"strings"
	"testing"
)

func TestRecommendedCandidatePlansForFixturesAreValid(t *testing.T) {
	for _, fixture := range loadToneFixtures(t) {
		t.Run(fixture.Name, func(t *testing.T) {
			plan, err := recommendedPlanForRig(fixture.Rig, false)
			if err != nil {
				t.Fatalf("recommendedPlanForRig() error = %v", err)
			}
			if plan.dsp[0] > helix.SafeDSPPerPath {
				t.Fatalf("path 0 DSP = %.1f%%, want at most %.1f%%", plan.dsp[0], helix.SafeDSPPerPath)
			}

			response := builderResponse{}
			for _, block := range plan.blocks {
				response.Blocks = append(response.Blocks, builderBlock{Name: block.name, ModelName: block.modelName, Path: block.path})
			}
			if err := validateBuilderResponse(response, &fixture.Rig, false); err != nil {
				t.Fatalf("planned response is invalid: %v", err)
			}
		})
	}
}

func TestRecommendedCandidatePlanRespectsDualDSPBudget(t *testing.T) {
	rig := RigDescription{Chain: []RigComponent{
		{Type: "amp", Name: "Amp", Description: "high gain modern"},
		{Type: "cab", Name: "Cab", Description: "4x12 modern"},
		{Type: "reverb", Name: "Reverb", Description: "large ambient stereo"},
	}}
	plan, err := recommendedPlanForRig(rig, true)
	if err != nil {
		t.Fatalf("recommendedPlanForRig() error = %v", err)
	}
	if plan.dsp[0] > helix.SafeDSPPerPath || plan.dsp[1] > helix.SafeDSPPerPath {
		t.Fatalf("plan DSP = %.1f%% / %.1f%%, want each at most %.1f%%", plan.dsp[0], plan.dsp[1], helix.SafeDSPPerPath)
	}
	if len(plan.blocks) != len(rig.Chain) {
		t.Fatalf("plan blocks = %d, want %d", len(plan.blocks), len(rig.Chain))
	}
}

func TestPlannerSelectsLowerRankCandidateWhenRequiredForDSP(t *testing.T) {
	groups := []rankedComponentCandidates{
		{component: RigComponent{Name: "Drive"}, candidates: []helix.CatalogEntry{{Name: "Preferred Drive", DSPMono: 50}, {Name: "Fallback Drive", DSPMono: 15}}},
		{component: RigComponent{Name: "Amp"}, candidates: []helix.CatalogEntry{{Name: "Amp", DSPMono: 50}}},
	}
	plan, err := recommendedPlanForGroups(groups, false)
	if err != nil {
		t.Fatalf("recommendedPlanForGroups() error = %v", err)
	}
	if got := plan.blocks[0].modelName; got != "Fallback Drive" {
		t.Fatalf("first planned model = %q, want DSP-safe fallback", got)
	}
}

func TestPlannerRejectsImpossibleSinglePathCombination(t *testing.T) {
	groups := []rankedComponentCandidates{{
		component:  RigComponent{Name: "Oversized Block"},
		candidates: []helix.CatalogEntry{{Name: "Oversized", DSPMono: helix.SafeDSPPerPath + 1}},
	}}
	_, err := recommendedPlanForGroups(groups, false)
	if err == nil || !strings.Contains(err.Error(), "Oversized Block") {
		t.Fatalf("recommendedPlanForGroups() error = %v, want impossible component error", err)
	}
}

func TestPlannerKeepsLaterBlocksOnSecondDSPPath(t *testing.T) {
	groups := []rankedComponentCandidates{
		{component: RigComponent{Name: "First"}, candidates: []helix.CatalogEntry{{Name: "First", DSPMono: 40}}},
		{component: RigComponent{Name: "Second"}, candidates: []helix.CatalogEntry{{Name: "Second", DSPMono: 40}}},
		{component: RigComponent{Name: "Third"}, candidates: []helix.CatalogEntry{{Name: "Third", DSPMono: 1}}},
	}
	plan, err := recommendedPlanForGroups(groups, true)
	if err != nil {
		t.Fatalf("recommendedPlanForGroups() error = %v", err)
	}
	for index, wantPath := range []int{0, 1, 1} {
		if got := plan.blocks[index].path; got != wantPath {
			t.Fatalf("block %d path = %d, want %d", index, got, wantPath)
		}
	}
}

func TestBuilderCorrectionInstructionRequiresCompleteReplacement(t *testing.T) {
	instruction := BuilderCorrectionInstruction(errors.New("invalid mapping"))
	for _, want := range []string{"complete replacement JSON object", "every non-Variax rig component exactly once", "recommended DSP-safe model/path tuples", "do not move later blocks back to path 0"} {
		if !strings.Contains(instruction, want) {
			t.Fatalf("BuilderCorrectionInstruction() does not contain %q", want)
		}
	}
}
