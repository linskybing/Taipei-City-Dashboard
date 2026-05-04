package aiassistant_test

import (
	"testing"

	"TaipeiCityDashboardBE/app/services/ai/assistant"
)

func TestContextDefaultsAndMetadata(t *testing.T) {
	ctx, err := assistant.NewContext("", "", "", " main-dashboard ")
	if err != nil {
		t.Fatalf("NewContext returned error: %v", err)
	}
	if ctx.Theme != assistant.DefaultTheme {
		t.Fatalf("Theme = %q, want %q", ctx.Theme, assistant.DefaultTheme)
	}
	if ctx.City != assistant.DefaultCity {
		t.Fatalf("City = %q, want %q", ctx.City, assistant.DefaultCity)
	}
	if ctx.Audience != assistant.DefaultAudience {
		t.Fatalf("Audience = %q, want %q", ctx.Audience, assistant.DefaultAudience)
	}
	if ctx.DashboardIndex != "main-dashboard" {
		t.Fatalf("DashboardIndex = %q, want trimmed value", ctx.DashboardIndex)
	}
	if ctx.Metadata()["dashboard_index"] != "main-dashboard" {
		t.Fatalf("metadata did not preserve dashboard index: %#v", ctx.Metadata())
	}
	if _, ok := ctx.Metadata()["decision_playbook"]; !ok {
		t.Fatalf("metadata missing decision playbook: %#v", ctx.Metadata())
	}
}

func TestContextAcceptsSixHackathonThemes(t *testing.T) {
	themes := []string{"auto", "commuting", "disaster", "environment", "health", "labor", "culture"}
	for _, theme := range themes {
		ctx, err := assistant.NewContext(theme, "taipei", "public", "")
		if err != nil {
			t.Fatalf("NewContext(%q) returned error: %v", theme, err)
		}
		if ctx.Theme != theme {
			t.Fatalf("Theme = %q, want %q", ctx.Theme, theme)
		}
	}
}

func TestContextRejectsInvalidValues(t *testing.T) {
	cases := []struct {
		name     string
		theme    string
		city     string
		audience string
	}{
		{name: "theme", theme: "sports", city: "taipei", audience: "public"},
		{name: "city", theme: "auto", city: "global", audience: "public"},
		{name: "audience", theme: "auto", city: "taipei", audience: "vendor"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := assistant.NewContext(tc.theme, tc.city, tc.audience, ""); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestThemePlaybooksCoverSixHackathonThemes(t *testing.T) {
	expectedQuestion := map[string]string{
		"commuting":   "哪個轉乘或最後一哩節點最可能失敗？",
		"disaster":    "未來數小時哪個避難點仍可到達且合適？",
		"environment": "何時何地的熱與空污暴露最低？",
		"health":      "近期食安事件是否影響家庭行動？",
		"labor":       "哪組訓練、就服與照顧資源走得完？",
		"culture":     "哪裡活動供給未接住新住民需求？",
	}

	for theme, question := range expectedQuestion {
		ctx, err := assistant.NewContext(theme, "metrotaipei", "government", "")
		if err != nil {
			t.Fatalf("NewContext(%q) returned error: %v", theme, err)
		}
		playbook, ok := ctx.Metadata()["decision_playbook"].(assistant.DecisionPlaybook)
		if !ok {
			t.Fatalf("decision_playbook type = %T", ctx.Metadata()["decision_playbook"])
		}
		if playbook.Question != question {
			t.Fatalf("%s question = %q, want %q", theme, playbook.Question, question)
		}
	}
}
