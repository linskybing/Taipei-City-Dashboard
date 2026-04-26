package ai

import (
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
	if maxToolLoops != 5 {
		t.Fatalf("maxToolLoops = %d, want 5", maxToolLoops)
	}
}
