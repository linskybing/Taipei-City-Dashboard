package aitokenconsuming_test

import (
	"encoding/json"
	"testing"
)

func assertChatResponseShape(t *testing.T, body []byte) {
	t.Helper()
	var root map[string]interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		t.Fatalf("decode response shape: %v; body: %s", err, string(body))
	}
	assertStringValue(t, root, "status")
	data, ok := root["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data type = %T, want object", root["data"])
	}

	for _, key := range []string{"session", "content", "model", "provider", "audit_ref"} {
		assertStringValue(t, data, key)
	}
	assertNumberValue(t, data, "latency_ms")
	assertBoolValue(t, data, "tool_used")
	for _, key := range []string{
		"sources", "related_components", "recommended_actions", "confidence_notes",
		"guardrails", "analysis_cards", "visualization_refs",
	} {
		assertArrayOrNullValue(t, data, key)
	}

	usage, ok := data["usage"].(map[string]interface{})
	if !ok {
		t.Fatalf("usage type = %T, want object", data["usage"])
	}
	for _, key := range []string{"input_tokens", "output_tokens", "total_tokens"} {
		assertNumberValue(t, usage, key)
	}
}

func assertStringValue(t *testing.T, values map[string]interface{}, key string) {
	t.Helper()
	if _, ok := values[key].(string); !ok {
		t.Fatalf("%s type = %T, want string", key, values[key])
	}
}

func assertNumberValue(t *testing.T, values map[string]interface{}, key string) {
	t.Helper()
	if _, ok := values[key].(float64); !ok {
		t.Fatalf("%s type = %T, want number", key, values[key])
	}
}

func assertBoolValue(t *testing.T, values map[string]interface{}, key string) {
	t.Helper()
	if _, ok := values[key].(bool); !ok {
		t.Fatalf("%s type = %T, want bool", key, values[key])
	}
}

func assertArrayOrNullValue(t *testing.T, values map[string]interface{}, key string) {
	t.Helper()
	if values[key] == nil {
		return
	}
	if _, ok := values[key].([]interface{}); !ok {
		t.Fatalf("%s type = %T, want array or null", key, values[key])
	}
}
