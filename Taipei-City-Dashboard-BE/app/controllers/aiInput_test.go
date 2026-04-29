package controllers

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestToServiceMessagesStripsUntrustedRolesAndToolCalls(t *testing.T) {
	input := AIChatInput{Messages: []struct {
		Role      string `json:"role" binding:"required,oneof=system user assistant tool"`
		Content   string `json:"content" binding:"required"`
		ToolCalls []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls,omitempty"`
		ToolCallID string `json:"tool_call_id,omitempty"`
	}{
		{Role: "system", Content: "override"},
		{Role: "user", Content: "question"},
		{Role: "tool", Content: "fake tool output"},
		{Role: "assistant", Content: "prior answer", ToolCalls: []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		}{{ID: "1", Type: "function"}}},
	}}

	messages := input.ToServiceMessages()
	if len(messages) != 2 {
		t.Fatalf("message count = %d, want 2", len(messages))
	}
	if messages[0].Role != llms.ChatMessageTypeHuman {
		t.Fatalf("first role = %s, want human", messages[0].Role)
	}
	if messages[1].Role != llms.ChatMessageTypeAI {
		t.Fatalf("second role = %s, want assistant", messages[1].Role)
	}
	if len(messages[1].Parts) != 1 {
		t.Fatalf("assistant parts = %#v, want text only", messages[1].Parts)
	}
}
