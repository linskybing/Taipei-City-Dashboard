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

func TestApplyContextPrioritizesComponentSearchIntent(t *testing.T) {
	ctx, err := NewContext("disaster", "metrotaipei", "government", "")
	if err != nil {
		t.Fatalf("NewContext returned error: %v", err)
	}
	messages := ApplyContext([]llms.MessageContent{{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: "想看防災相關組件"}},
	}}, ctx)
	text := firstTextPart(messages[0])
	for _, want := range []string{
		"查看組件",
		"推薦組件",
		"想看某議題資料",
		"優先呼叫 search_components",
		"再呼叫 get_component_snapshot",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("system instruction missing %q:\n%s", want, text)
		}
	}
}

func TestApplyContextRequiresOpenDataThroughAirflow(t *testing.T) {
	ctx, err := NewContext("commuting", "metrotaipei", "government", "")
	if err != nil {
		t.Fatalf("NewContext returned error: %v", err)
	}
	messages := ApplyContext([]llms.MessageContent{{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: "請加入新的開放資料 API 做捷運異常分析"}},
	}}, ctx)
	text := firstTextPart(messages[0])
	for _, want := range []string{
		"Data-End Airflow DAG",
		"PostgreSQL",
		"不得建議或執行前端、後端工具直接串接",
		"CommonDag ETL",
		"job_config metadata",
		"component query",
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
