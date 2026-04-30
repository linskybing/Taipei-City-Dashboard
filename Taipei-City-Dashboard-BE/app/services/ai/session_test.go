package ai

import (
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/global"
	"context"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
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

func TestShouldLoadMemoryOnlyForLatestMessageRequests(t *testing.T) {
	req := AIChatRequest{
		SessionID: "session_1",
		UserID:    "7",
		Messages: []llms.MessageContent{{
			Role: llms.ChatMessageTypeHuman,
		}},
	}
	if !shouldLoadMemory(req) {
		t.Fatal("expected memory for single-message request")
	}
	req.Messages = append(req.Messages, llms.MessageContent{Role: llms.ChatMessageTypeHuman})
	if shouldLoadMemory(req) {
		t.Fatal("did not expect memory when caller already sends history")
	}
}

func TestBuildSessionMemoryBoundsContent(t *testing.T) {
	longAnswer := strings.Repeat("資料判讀很長 ", 80)
	memory := buildSessionMemory([]models.AIChatLog{
		{Question: "第一題\n含換行", Answer: "第一答"},
		{Question: "第二題", Answer: longAnswer},
	})
	for _, want := range []string{"對話記憶", "1. Q:第一題 含換行 A:第一答", "2. Q:第二題 A:"} {
		if !strings.Contains(memory, want) {
			t.Fatalf("memory missing %q: %s", want, memory)
		}
	}
	if got := len([]rune(memory)); got > maxMemoryBlockRunes {
		t.Fatalf("memory length = %d, want <= %d", got, maxMemoryBlockRunes)
	}
}

func TestTrimMemoryTextUsesRuneLimitWithEllipsis(t *testing.T) {
	text := strings.Repeat("臺北資料", 20)
	trimmed := trimMemoryText(text, 10)
	if got := len([]rune(trimmed)); got != 10 {
		t.Fatalf("trimmed length = %d, want 10: %q", got, trimmed)
	}
	if !strings.HasSuffix(trimmed, memoryEllipsis) {
		t.Fatalf("trimmed text should end with ellipsis: %q", trimmed)
	}
	if strings.Contains(trimMemoryText("keep", 0), "keep") {
		t.Fatal("zero limit should not retain content")
	}
}
