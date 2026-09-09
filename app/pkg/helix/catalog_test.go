package helix

import "testing"

func TestCandidatesForComponentAreTypeCompatible(t *testing.T) {
	for _, componentType := range []string{"amp", "cab", "pedal", "modulation", "delay", "reverb"} {
		candidates := CandidatesForComponent(componentType)
		if len(candidates) == 0 {
			t.Fatalf("CandidatesForComponent(%q) returned no candidates", componentType)
		}
		for _, candidate := range candidates {
			if !IsCompatibleComponentModel(componentType, candidate) {
				t.Fatalf("candidate %q is not compatible with %q", candidate.InternalName, componentType)
			}
		}
	}
}

func TestCandidatesForComponentRejectsOtherFamilies(t *testing.T) {
	delay := CandidatesForComponent("delay")[0]
	if IsCompatibleComponentModel("amp", delay) {
		t.Fatalf("delay %q must not be accepted as an amp", delay.InternalName)
	}
}

func TestEffectiveDSPMonoUsesFallback(t *testing.T) {
	if got := EffectiveDSPMono(CatalogEntry{}); got != DefaultDSPMono {
		t.Fatalf("EffectiveDSPMono() = %.1f, want fallback %.1f", got, DefaultDSPMono)
	}
	if got := EffectiveDSPMono(CatalogEntry{DSPMono: 12.5}); got != 12.5 {
		t.Fatalf("EffectiveDSPMono() = %.1f, want catalog cost", got)
	}
}

func TestEffectiveDSPUsesStereoDefaults(t *testing.T) {
	entry := CatalogEntry{
		DSPMono:   3,
		DSPStereo: 5,
		Data: map[string]interface{}{
			"Defaults": map[string]interface{}{"@stereo": true},
		},
	}
	if got := EffectiveDSP(entry); got != 5 {
		t.Fatalf("EffectiveDSP() = %.1f, want stereo cost 5.0", got)
	}
	entry.Data["Defaults"].(map[string]interface{})["@stereo"] = false
	if got := EffectiveDSP(entry); got != 3 {
		t.Fatalf("EffectiveDSP() = %.1f, want mono cost 3.0", got)
	}
}

func TestRankCandidatesPrioritizesToneIntent(t *testing.T) {
	candidates := []CatalogEntry{
		{InternalName: "HD2_AmpUSDeluxe", Name: "US Deluxe", BasedOn: "Fender Deluxe Reverb"},
		{InternalName: "HD2_AmpCaliRectifire", Name: "Cali Rectifire", BasedOn: "Mesa/Boogie Dual Rectifier"},
	}

	ranked := rankCandidates(candidates, "metal moderno de alta ganancia", 2)
	if ranked[0].InternalName != "HD2_AmpCaliRectifire" {
		t.Fatalf("top candidate = %q, want high-gain amp", ranked[0].InternalName)
	}
}

func TestRankedCandidatesForComponentLimitsCompatibleModels(t *testing.T) {
	ranked := RankedCandidatesForComponent("amp", "clean vintage blues", PromptCandidateLimit)
	if len(ranked) != PromptCandidateLimit {
		t.Fatalf("ranked amp count = %d, want %d", len(ranked), PromptCandidateLimit)
	}
	for _, candidate := range ranked {
		if !IsCompatibleComponentModel("amp", candidate) {
			t.Fatalf("ranked candidate %q is not an amp", candidate.InternalName)
		}
	}
}
