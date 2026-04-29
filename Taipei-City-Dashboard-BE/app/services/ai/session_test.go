package ai

import (
	"TaipeiCityDashboardBE/global"
	"context"
	"strings"
	"testing"
)

func TestExecuteAllowedToolRejectsUnlistedTool(t *testing.T) {
	session := &aiSession{allowedTools: map[string]bool{"search_components": true}}

	_, err := session.executeAllowedTool(context.Background(), "get_current_time", "{}")
	if err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("expected unlisted tool rejection, got %v", err)
	}
}

func TestMaxToolLoopsBounded(t *testing.T) {
	original := global.TWCC.MaxToolLoops
	defer func() { global.TWCC.MaxToolLoops = original }()

	global.TWCC.MaxToolLoops = 0
	if configuredToolLoops() != 1 {
		t.Fatalf("configuredToolLoops = %d, want 1", configuredToolLoops())
	}
	global.TWCC.MaxToolLoops = 50
	if configuredToolLoops() != maxAllowedToolLoops {
		t.Fatalf("configuredToolLoops = %d, want %d", configuredToolLoops(), maxAllowedToolLoops)
	}
}

func TestNormalizeToolArgsRequiresJSONObject(t *testing.T) {
	for _, args := range []string{`[]`, `"text"`, `{bad-json`} {
		if _, err := normalizeToolArgs(args); err == nil {
			t.Fatalf("expected invalid args error for %q", args)
		}
	}
	normalized, err := normalizeToolArgs("")
	if err != nil {
		t.Fatalf("empty args should normalize to object: %v", err)
	}
	if normalized != "{}" {
		t.Fatalf("normalized = %q, want {}", normalized)
	}
}

func TestApplyToolContextDefaults(t *testing.T) {
	session := &aiSession{req: AIChatRequest{Params: map[string]interface{}{
		"theme": "health", "city": "taipei", "audience": "public",
	}}}
	args := session.applyToolContextDefaults(`{"component_id":12}`)
	for _, expected := range []string{`"component_id":12`, `"theme":"health"`, `"city":"taipei"`, `"audience":"public"`} {
		if !strings.Contains(args, expected) {
			t.Fatalf("args %s missing %s", args, expected)
		}
	}
}

func TestToolResultStatusDistinguishesUnavailable(t *testing.T) {
	result := `{"tool":"x","data":{"status":"unavailable"}}`
	if status := toolResultStatus(result); status != "unavailable" {
		t.Fatalf("status = %s, want unavailable", status)
	}
	if status := toolResultStatus(`{"tool":"x"}`); status != "success" {
		t.Fatalf("status = %s, want success", status)
	}
}
