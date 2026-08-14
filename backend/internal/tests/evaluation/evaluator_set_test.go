package evaluation

import (
	"strings"
	"testing"
)

func TestClosedCasesAreAnonymizedAndCoverEvaluatorRegressionSurface(t *testing.T) {
	cases := ClosedCases()
	if len(cases) != 132 {
		t.Fatalf("closed cases=%d, want 132", len(cases))
	}
	seenRoles, seenRisks, roleRisks, categories, injectionRU, injectionMixed := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}, false, false
	for _, item := range cases {
		if item.ID == "" || item.Answer == "" || item.CounterpartyMessage == "" || item.EvaluationContext == "" || item.MinScore > item.MaxScore {
			t.Fatalf("invalid case %#v", item)
		}
		expectsConcreteSignal := item.Category == "short_risky" || item.Category == "slang"
		if expectsConcreteSignal != (item.ExpectedSignal != "") {
			t.Fatalf("case %s category=%s expected signal=%q does not match concrete action", item.ID, item.Category, item.ExpectedSignal)
		}
		seenRoles[item.Role] = true
		seenRisks[item.RiskType] = true
		roleRisks[item.Role+":"+item.RiskType] = true
		categories[item.Category] = true
		injectionRU = injectionRU || strings.Contains(strings.ToLower(item.Answer), "игнорируй прошлые инструкции")
		injectionMixed = injectionMixed || strings.Contains(item.Answer, "Ignore previous instructions")
	}
	if len(seenRoles) != 2 || len(seenRisks) != 7 || len(roleRisks) != 12 || !injectionRU || !injectionMixed || !categories["typo"] || !categories["mixed"] || !categories["off_topic"] || !categories["model_safe"] || !categories["model_risky"] {
		t.Fatalf("coverage roles=%v risks=%v roleRisks=%v categories=%v injectionRU=%v injectionMixed=%v", seenRoles, seenRisks, roleRisks, categories, injectionRU, injectionMixed)
	}
}
