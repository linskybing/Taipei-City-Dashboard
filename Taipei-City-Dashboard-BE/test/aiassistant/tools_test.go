package aiassistant_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"TaipeiCityDashboardBE/app/services/ai/assistant"
	"TaipeiCityDashboardBE/app/services/ai/tools"
	"github.com/tmc/langchaingo/llms"
)

func TestToolDefinitionsExposeOnlyAssistantTools(t *testing.T) {
	expected := map[string]bool{
		"search_components":       true,
		"get_component_snapshot":  true,
		"compare_city_components": true,
		"get_dashboard_context":   true,
		"recommend_actions":       true,
	}

	for _, tool := range assistant.ToolDefinitions() {
		if tool.Function == nil {
			t.Fatalf("tool has nil function definition: %#v", tool)
		}
		name := tool.Function.Name
		if !expected[name] {
			t.Fatalf("unexpected tool definition %q", name)
		}
		delete(expected, name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing tool definitions: %#v", expected)
	}
}

func TestToolDefinitionsCarrySchemaEnums(t *testing.T) {
	definitions := assistant.ToolDefinitions()
	search := findTool(t, definitions, "search_components")
	params, ok := search.Function.Parameters.(map[string]interface{})
	if !ok {
		t.Fatalf("parameters type = %T, want map", search.Function.Parameters)
	}
	properties := params["properties"].(map[string]interface{})
	required := params["required"].([]string)

	if !contains(required, "query") {
		t.Fatalf("search_components required = %#v, want query", required)
	}
	assertEnum(t, properties["theme"], "commuting", "disaster", "environment", "health", "labor", "culture")
	assertEnum(t, properties["city"], "taipei", "metrotaipei")
}

func TestToolExecutionValidation(t *testing.T) {
	if _, err := tools.Execute(context.Background(), "unknown_tool", "{}"); err == nil {
		t.Fatal("expected unknown tool error")
	}

	output, err := tools.Execute(context.Background(), "recommend_actions",
		`{"theme":"health","city":"taipei","audience":"public","signals":["食品稽查"]}`)
	if err != nil {
		t.Fatalf("recommend_actions returned error: %v", err)
	}
	var envelope map[string]interface{}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("invalid tool JSON: %v", err)
	}
	if envelope["tool"] != "recommend_actions" {
		t.Fatalf("tool = %v, want recommend_actions", envelope["tool"])
	}
	if actions, ok := envelope["recommended_actions"].([]interface{}); !ok || len(actions) == 0 {
		t.Fatalf("recommended_actions missing from output: %s", output)
	}
}

func TestRecommendActionsRejectsBadInput(t *testing.T) {
	cases := []string{
		`{bad-json`,
		`{"theme":"auto","city":"global","audience":"public"}`,
	}
	for _, args := range cases {
		_, err := tools.Execute(context.Background(), "recommend_actions", args)
		if err == nil || strings.TrimSpace(err.Error()) == "" {
			t.Fatalf("expected error for args %q", args)
		}
	}
}

func findTool(t *testing.T, tools []llms.Tool, name string) llms.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Function != nil && tool.Function.Name == name {
			return tool
		}
	}
	t.Fatalf("tool %q not found", name)
	return llms.Tool{}
}
