package evaluation

import (
	"strings"
	"testing"
)

func TestClosedCasesAreAnonymizedAndCoverEvaluatorRegressionSurface(t *testing.T) {
	cases := ClosedCases()
	if len(cases) < 100 || len(cases) > 200 {
		t.Fatalf("closed cases=%d, want 100..200", len(cases))
	}
	seenRoles, seenRisks, categories, injectionRU, injectionMixed := map[string]bool{}, map[string]bool{}, map[string]bool{}, false, false
	for _, item := range cases {
		if item.ID == "" || item.Answer == "" || item.ExpectedSignal == "" || item.MinScore > item.MaxScore {
			t.Fatalf("invalid case %#v", item)
		}
		seenRoles[item.Role] = true
		seenRisks[item.RiskType] = true
		categories[item.Category] = true
		injectionRU = injectionRU || strings.Contains(item.Answer, "игнорируй прошлые инструкции")
		injectionMixed = injectionMixed || strings.Contains(item.Answer, "Ignore previous instructions")
	}
	if len(seenRoles) != 2 || len(seenRisks) != 6 || !injectionRU || !injectionMixed || !categories["typo"] || !categories["mixed"] || !categories["off_topic"] || !categories["same_wording"] {
		t.Fatalf("coverage roles=%v risks=%v categories=%v injectionRU=%v injectionMixed=%v", seenRoles, seenRisks, categories, injectionRU, injectionMixed)
	}
}
