package domain_test

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"testing"
)

func TestScenarioPublicationDataRequiresSupportedRiskAndProductContext(t *testing.T) {
	valid := domain.ProductContext{
		ItemTitle:  "Смартфон Apple iPhone 15",
		Category:   "Электроника",
		DealMethod: "delivery",
		Price:      67000,
		Currency:   "RUB",
		Location:   "Москва",
		ImageKey:   "smartphone",
	}
	if !domain.ValidProductContext(valid) {
		t.Fatalf("complete product context rejected: %#v", valid)
	}
	for name, mutate := range map[string]func(*domain.ProductContext){
		"missing title":           func(value *domain.ProductContext) { value.ItemTitle = "" },
		"missing category":        func(value *domain.ProductContext) { value.Category = "" },
		"unsupported deal method": func(value *domain.ProductContext) { value.DealMethod = "messenger" },
		"remote image":            func(value *domain.ProductContext) { value.ImageKey = "https://example.com/item.png" },
		"price without currency":  func(value *domain.ProductContext) { value.Currency = "" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := valid
			mutate(&candidate)
			if domain.ValidProductContext(candidate) {
				t.Fatalf("invalid product context accepted: %#v", candidate)
			}
		})
	}

	for _, risk := range []domain.RiskType{
		domain.RiskPhishing,
		domain.RiskPrepayment,
		domain.RiskFakePayment,
		domain.RiskDelivery,
		domain.RiskExternalMessenger,
		domain.RiskAccountTakeover,
		domain.RiskSMSCode,
		domain.RiskSocialEngineering,
	} {
		if !domain.ValidRiskType(risk) {
			t.Errorf("documented risk type %q rejected", risk)
		}
	}
	if domain.ValidRiskType("other") {
		t.Fatal("unsupported risk type accepted")
	}
}

func TestAssessmentCategoriesFollowAllowedPoints(t *testing.T) {
	for points, want := range map[int]string{0: "unsafe", 25: "unsafe", 50: "risky", 75: "mostly_safe", 100: "safe"} {
		if got := domain.AssessmentForPoints(points); got != want {
			t.Errorf("AssessmentForPoints(%d)=%q, want %q", points, got, want)
		}
	}
	for _, points := range []int{26, 76, 101} {
		if got := domain.AssessmentForPoints(points); got != "invalid" {
			t.Errorf("AssessmentForPoints(%d)=%q, want invalid", points, got)
		}
	}
}
