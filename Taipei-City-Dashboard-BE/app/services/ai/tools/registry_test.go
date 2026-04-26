package tools

import (
	"context"
	"strings"
	"testing"
)

func TestExecuteRejectsUnknownTool(t *testing.T) {
	_, err := Execute(context.Background(), "unknown_tool", "{}")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected unknown tool error, got %v", err)
	}
}
