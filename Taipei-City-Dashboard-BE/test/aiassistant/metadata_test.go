package aiassistant_test

import (
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
			"confidence_notes":["Qdrant metadata match"]
		}`,
		`{
			"tool":"recommend_actions",
			"sources":[{"name":"data.taipei","type":"open_data","url":"https://data.taipei"}],
			"related_components":[{"id":1,"index":"air_quality","name":"空氣品質","city":"taipei","score":0.89}],
			"recommended_actions":["檢視污染熱點","安排跨局處追蹤"],
			"confidence_notes":["Qdrant metadata match","需以業務資料確認"]
		}`,
	}

	raw := assistant.BuildMetadata(
		map[string]interface{}{"theme": "environment", "city": "taipei", "audience": "government"},
		[]string{"search_components", "recommend_actions"},
		outputs,
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
}

func TestResponseExtrasIgnoreInvalidMetadata(t *testing.T) {
	extras := assistant.ResponseExtrasFromMetadata(`{bad-json`)
	if len(extras.Sources) != 0 || len(extras.RelatedComponents) != 0 {
		t.Fatalf("expected empty extras, got %#v", extras)
	}
}
