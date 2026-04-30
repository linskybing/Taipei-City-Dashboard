package assistant

import (
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

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

func TestApplyContextIncludesStatRouting(t *testing.T) {
	ctx, err := NewContext("environment", "taipei", "government", "")
	if err != nil {
		t.Fatalf("NewContext returned error: %v", err)
	}
	messages := ApplyContext([]llms.MessageContent{{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: "component_id 1 完整統計診斷"}},
	}}, ctx)
	if len(messages) == 0 || messages[0].Role != llms.ChatMessageTypeSystem {
		t.Fatalf("missing system message: %#v", messages)
	}
	text := firstTextPart(messages[0])
	for _, want := range []string{
		"完整統計診斷",
		"component_id",
		"clean_impute",
		"descriptive_report",
		"trend_detect",
		"seasonal_decompose",
		"anomaly_detect",
		"hypothesis_test",
		"forecast_short_mid",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("system instruction missing %q:\n%s", want, text)
		}
	}
}

func firstTextPart(msg llms.MessageContent) string {
	for _, part := range msg.Parts {
		text, ok := part.(llms.TextContent)
		if ok {
			return text.Text
		}
	}
	return ""
}
