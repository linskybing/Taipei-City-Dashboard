package assistant

import "testing"

func TestNewContextDefaults(t *testing.T) {
	ctx, err := NewContext("", "", "", "")
	if err != nil {
		t.Fatalf("NewContext returned error: %v", err)
	}
	if ctx.Theme != DefaultTheme || ctx.City != DefaultCity || ctx.Audience != DefaultAudience {
		t.Fatalf("unexpected defaults: %+v", ctx)
	}
}

func TestNewContextRejectsInvalidTheme(t *testing.T) {
	if _, err := NewContext("invalid", "taipei", "government", ""); err == nil {
		t.Fatal("expected invalid theme error")
	}
}

func TestRecommendActionsRejectsInvalidCity(t *testing.T) {
	_, err := RecommendActions(nil, `{"theme":"auto","city":"global","audience":"public"}`)
	if err == nil {
		t.Fatal("expected invalid city error")
	}
}

func TestRecommendActionsRejectsInvalidJSON(t *testing.T) {
	_, err := RecommendActions(nil, `{bad-json`)
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
