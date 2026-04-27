package ai

import (
	"TaipeiCityDashboardBE/app/services/ai/assistant"
	"TaipeiCityDashboardBE/app/services/ai/tools"
	"TaipeiCityDashboardBE/global"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (s *aiSession) executeAllowedTool(ctx context.Context, name string, args string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("tool name is required")
	}
	if !s.allowedTools[name] {
		return "", fmt.Errorf("tool %s is not allowed", name)
	}
	normalizedArgs, err := normalizeToolArgs(args)
	if err != nil {
		return "", err
	}
	toolCtx, cancel := context.WithTimeout(ctx, time.Duration(configuredToolTimeout())*time.Second)
	defer cancel()
	return tools.Execute(toolCtx, name, normalizedArgs)
}

func (s *aiSession) recordToolEvent(name, status string, loop, latencyMS int, err error) {
	event := assistant.ToolEvent{
		Name:      name,
		Status:    status,
		Loop:      loop,
		LatencyMS: latencyMS,
	}
	if err != nil {
		event.Error = err.Error()
	}
	s.toolEvents = append(s.toolEvents, event)
}

func normalizeToolArgs(args string) (string, error) {
	trimmed := strings.TrimSpace(args)
	if trimmed == "" {
		trimmed = "{}"
	}
	if len([]byte(trimmed)) > configuredToolArgBytes() {
		return "", fmt.Errorf("tool arguments exceed size limit")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return "", fmt.Errorf("tool arguments must be a JSON object: %v", err)
	}
	if payload == nil {
		return "", fmt.Errorf("tool arguments must be a JSON object")
	}
	return trimmed, nil
}

func buildToolFallback(name string, err error) string {
	name = auditToolName(name)
	payload := map[string]interface{}{
		"tool": name,
		"confidence_notes": []string{
			fmt.Sprintf("工具 %s 暫時無法取得資料，回答需標示資料限制。", name),
		},
		"data": map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		},
	}
	raw, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return fmt.Sprintf(`{"tool":%q,"data":{"status":"error"}}`, name)
	}
	return string(raw)
}

func auditToolName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "unknown_tool"
	}
	return trimmed
}

func configuredToolLoops() int {
	if global.TWCC.MaxToolLoops < 1 {
		return 1
	}
	if global.TWCC.MaxToolLoops > maxAllowedToolLoops {
		return maxAllowedToolLoops
	}
	return global.TWCC.MaxToolLoops
}

func configuredToolTimeout() int {
	if global.TWCC.ToolTimeout < 1 {
		return 1
	}
	return global.TWCC.ToolTimeout
}

func configuredToolArgBytes() int {
	if global.TWCC.MaxToolArgBytes < 2 {
		return 2
	}
	return global.TWCC.MaxToolArgBytes
}
