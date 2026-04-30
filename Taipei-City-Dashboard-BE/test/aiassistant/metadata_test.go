package aiassistant_test

import (
	"encoding/json"
	"testing"

	"TaipeiCityDashboardBE/app/services/ai/assistant"
)

func TestMetadataMergesToolOutputs(t *testing.T) {
	outputs := []string{
		`{
			"tool":"search_components",
			"sources":[{"name":"data.taipei","type":"open_data","url":"https://data.taipei"}],
			"related_components":[{"id":1,"index":"air_quality","name":"空氣品質","city":"taipei","score":0.91}],
			"recommended_actions":["檢視污染熱點"],
			"confidence_notes":["Qdrant metadata match"],
			"guardrails":["只使用回傳組件"]
		}`,
		`{
			"tool":"recommend_actions",
			"sources":[{"name":"data.taipei","type":"open_data","url":"https://data.taipei"}],
			"related_components":[{"id":1,"index":"air_quality","name":"空氣品質","city":"taipei","score":0.89}],
			"recommended_actions":["檢視污染熱點","安排跨局處追蹤"],
			"confidence_notes":["Qdrant metadata match","需以業務資料確認"],
			"guardrails":["只使用回傳組件","需要業務確認"],
			"analysis_cards":[{"tool":"descriptive_report","headline":"描述性統計已產生","confidence_label":"high"}]
		}`,
	}

	raw := assistant.BuildMetadata(
		map[string]interface{}{"theme": "environment", "city": "taipei", "audience": "government"},
		[]string{"search_components", "recommend_actions"},
		outputs,
		[]assistant.ToolEvent{{Name: "search_components", Status: "success", Loop: 1, LatencyMS: 12}},
	)
	extras := assistant.ResponseExtrasFromMetadata(raw)

	if len(extras.Sources) != 1 {
		t.Fatalf("Sources length = %d, want 1", len(extras.Sources))
	}
	if len(extras.RelatedComponents) != 1 {
		t.Fatalf("RelatedComponents length = %d, want 1", len(extras.RelatedComponents))
	}
	if len(extras.RecommendedActions) != 2 {
		t.Fatalf("RecommendedActions length = %d, want 2", len(extras.RecommendedActions))
	}
	if len(extras.ConfidenceNotes) != 2 {
		t.Fatalf("ConfidenceNotes length = %d, want 2", len(extras.ConfidenceNotes))
	}
	if len(extras.Guardrails) != 2 {
		t.Fatalf("Guardrails length = %d, want 2", len(extras.Guardrails))
	}
	if len(extras.AnalysisCards) != 1 {
		t.Fatalf("AnalysisCards length = %d, want 1", len(extras.AnalysisCards))
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		t.Fatalf("metadata JSON invalid: %v", err)
	}
	events, ok := metadata["tool_events"].([]interface{})
	if !ok || len(events) != 1 {
		t.Fatalf("tool_events missing from metadata: %s", raw)
	}
}

func TestResponseExtrasIgnoreInvalidMetadata(t *testing.T) {
	extras := assistant.ResponseExtrasFromMetadata(`{bad-json`)
	if len(extras.Sources) != 0 || len(extras.RelatedComponents) != 0 {
		t.Fatalf("expected empty extras, got %#v", extras)
	}
}
