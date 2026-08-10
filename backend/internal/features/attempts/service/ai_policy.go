package service

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed content/policies.json
var compactPoliciesJSON []byte

var compactPolicies = func() map[string]map[string]string {
	var policies map[string]map[string]string
	if err := json.Unmarshal(compactPoliciesJSON, &policies); err != nil {
		panic("invalid embedded AI policy: " + err.Error())
	}
	return policies
}()

func PolicyFor(role, riskType string) string {
	rolePolicies := compactPolicies[role]
	risk := rolePolicies[riskType]
	if risk == "" {
		risk = rolePolicies["social_engineering"]
	}
	return strings.TrimSpace(rolePolicies["_core"] + " " + risk)
}
