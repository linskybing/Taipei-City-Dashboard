package assistant

import (
	"encoding/json"
	"testing"
)

func TestBuildMetadataMergesToolOutputs(t *testing.T) {
	output, err := marshalTool(toolEnvelope{
		Tool:               "recommend_actions",
		Sources:            []Source{{Name: "data.taipei", URL: "https://data.taipei"}},
		RelatedComponents:  []RelatedComponent{{ID: 1, Index: "demo", City: "taipei"}},
		RecommendedActions: []string{"action"},
		ConfidenceNotes:    []string{"note"},
		Guardrails:         []string{"guardrail"},
	})
	if err != nil {
		t.Fatalf("marshalTool returned error: %v", err)
	}

	raw := BuildMetadata(
		map[string]interface{}{"theme": "auto"},
		[]string{"recommend_actions"},
		[]string{output},
		[]ToolEvent{{Name: "recommend_actions", Status: "success", Loop: 1}},
	)
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		t.Fatalf("metadata is not JSON: %v", err)
	}

	extras := ResponseExtrasFromMetadata(raw)
	if len(extras.Sources) != 1 || len(extras.RelatedComponents) != 1 {
		t.Fatalf("unexpected extras: %+v", extras)
	}
	if extras.RecommendedActions[0] != "action" || extras.ConfidenceNotes[0] != "note" {
		t.Fatalf("missing action or confidence note: %+v", extras)
	}
	if len(extras.Guardrails) != 1 || extras.Guardrails[0] != "guardrail" {
		t.Fatalf("missing guardrail: %+v", extras)
	}
	if events, ok := metadata["tool_events"].([]interface{}); !ok || len(events) != 1 {
		t.Fatalf("missing tool events: %s", raw)
	}
}
