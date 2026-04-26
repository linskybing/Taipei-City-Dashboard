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
