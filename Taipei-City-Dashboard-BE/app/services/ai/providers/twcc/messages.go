package twcc

import "github.com/tmc/langchaingo/llms"

func (m *TWCC) toTWCCMessages(messages []llms.MessageContent) []TWCCMessage {
	twccMessages := make([]TWCCMessage, 0, len(messages))
	for _, mc := range messages {
		role := m.mapRole(mc.Role)
		var toolCallID, toolName string
		var twccToolCalls []TWCCToolCall
		var contentText string

		for _, part := range mc.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				contentText = cleanXML(p.Text)
			case llms.ToolCall:
				if p.FunctionCall == nil {
					continue
				}
				twccToolCalls = append(twccToolCalls, TWCCToolCall{
					ID:   p.ID,
					Type: p.Type,
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      p.FunctionCall.Name,
						Arguments: cleanXML(p.FunctionCall.Arguments),
					},
				})
			case llms.ToolCallResponse:
				toolCallID = p.ToolCallID
				toolName = p.Name
				contentText = cleanXML(p.Content)
			}
		}

		msg := TWCCMessage{
			Role:       role,
			ToolCalls:  twccToolCalls,
			ToolCallID: toolCallID,
			Name:       toolName,
		}
		if role == "assistant" && len(twccToolCalls) > 0 && contentText == "" {
			msg.Content = nil
		} else {
			msg.Content = strPtr(contentText)
		}
		twccMessages = append(twccMessages, msg)
	}
	return twccMessages
}

func (m *TWCC) mapRole(role llms.ChatMessageType) string {
	switch role {
	case llms.ChatMessageTypeHuman:
		return "user"
	case llms.ChatMessageTypeAI:
		return "assistant"
	case llms.ChatMessageTypeSystem:
		return "system"
	case llms.ChatMessageTypeTool:
		return "tool"
	default:
		return string(role)
	}
}

func strPtr(s string) *string {
	return &s
}
