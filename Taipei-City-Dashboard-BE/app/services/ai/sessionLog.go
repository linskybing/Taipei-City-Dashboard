package ai

import (
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/app/services/ai/assistant"
	"TaipeiCityDashboardBE/global"
	"TaipeiCityDashboardBE/logs"
	"context"
	"encoding/json"
	"time"

	"github.com/tmc/langchaingo/llms"
)

func (s *aiSession) finalize(ctx context.Context) (*models.AIChatLog, error) {
	log := &models.AIChatLog{
		SessionID:    s.req.SessionID,
		UserID:       s.req.UserID,
		IPAddress:    s.req.IPAddress,
		Provider:     "twcc",
		Model:        global.TWCC.Model,
		LatencyMS:    int(time.Since(s.startTime).Milliseconds()),
		Status:       "success",
		Tools:        "[]",
		CreatedAt:    s.startTime,
		Metadata:     assistant.BuildMetadata(s.metadataParams(), s.executedTools, s.toolResults, s.toolEvents),
		InputTokens:  s.totalInput,
		OutputTokens: s.totalOutput,
		TotalTokens:  s.totalInput + s.totalOutput,
	}
	if len(s.req.Messages) > 0 {
		log.Question = extractText(s.req.Messages[len(s.req.Messages)-1])
	}
	if s.lastErr != nil {
		log.Status = "error"
		log.ErrorCode = "MODEL_ERROR"
		log.ErrorMessage = s.lastErr.Error()
		createAIChatLog(ctx, log)
		return log, s.lastErr
	}
	if s.lastResp != nil && len(s.lastResp.Choices) > 0 {
		log.Answer = s.lastResp.Choices[0].Content
		if s.toolUsed {
			log.ToolUsed = true
			if toolJSON, err := json.Marshal(s.executedTools); err == nil {
				log.Tools = string(toolJSON)
			}
		}
	}
	if err := createAIChatLog(ctx, log); err != nil {
		logs.FError("DB Log Error: %v", err)
	}
	return log, nil
}

func toolsToParts(calls []llms.ToolCall) []llms.ContentPart {
	parts := make([]llms.ContentPart, len(calls))
	for i, call := range calls {
		parts[i] = call
	}
	return parts
}

func toolCallPayload(call llms.ToolCall) (string, string) {
	if call.FunctionCall == nil {
		return "", ""
	}
	return call.FunctionCall.Name, call.FunctionCall.Arguments
}

func allowedToolMap(tools []llms.Tool) map[string]bool {
	allowed := make(map[string]bool, len(tools))
	for _, tool := range tools {
		if tool.Function != nil {
			allowed[tool.Function.Name] = true
		}
	}
	return allowed
}

func toolNames(tools []llms.Tool) string {
	names := ""
	count := 0
	for _, tool := range tools {
		if tool.Function == nil {
			continue
		}
		if count > 0 {
			names += ", "
		}
		names += tool.Function.Name
		count++
	}
	return names
}

func mergeSystemMsg(msg llms.MessageContent, instruction string) llms.MessageContent {
	newParts := make([]llms.ContentPart, len(msg.Parts))
	for i, part := range msg.Parts {
		if text, ok := part.(llms.TextContent); ok {
			newParts[i] = llms.TextContent{Text: text.Text + instruction}
			continue
		}
		newParts[i] = part
	}
	return llms.MessageContent{Role: msg.Role, Parts: newParts}
}

func extractText(msg llms.MessageContent) string {
	for _, part := range msg.Parts {
		if text, ok := part.(llms.TextContent); ok {
			return text.Text
		}
	}
	return ""
}

func parseUsageInt(val interface{}) int {
	switch value := val.(type) {
	case int:
		return value
	case float64:
		return int(value)
	default:
		return 0
	}
}
