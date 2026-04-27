package twcc

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"TaipeiCityDashboardBE/logs"
	"github.com/tmc/langchaingo/llms"
)

func (m *TWCC) handleStandardResponse(body io.Reader) (*llms.ContentResponse, error) {
	raw, _ := io.ReadAll(body)
	logs.FInfo("TWCC Raw Response: %s", string(raw))

	var tr TWCCResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %v", err)
	}

	content := tr.GeneratedText
	tools := toolCallsFromResponse(tr)
	if len(tr.Choices) > 0 && tr.Choices[0].Message.Content != "" {
		content = tr.Choices[0].Message.Content
	}
	if len(tools) == 0 && looksLikeXMLTool(content) {
		tools, content = extractXMLToolCalls(content)
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{
			Content: content,
			GenerationInfo: map[string]interface{}{
				"model":      m.ModelName,
				"tool_calls": tools,
				"usage": map[string]interface{}{
					"input_tokens":  tr.PromptTokens,
					"output_tokens": tr.GeneratedTokens,
					"total_tokens":  tr.TotalTokens,
				},
			},
		}},
	}, nil
}

func toolCallsFromResponse(resp TWCCResponse) []llms.ToolCall {
	if len(resp.ToolCalls) > 0 {
		return toolCallsFromTWCC(resp.ToolCalls)
	}
	if len(resp.Choices) == 0 {
		return nil
	}
	return toolCallsFromTWCC(resp.Choices[0].Message.ToolCalls)
}

func toolCallsFromTWCC(calls []TWCCToolCall) []llms.ToolCall {
	tools := make([]llms.ToolCall, 0, len(calls))
	for _, call := range calls {
		tools = append(tools, llms.ToolCall{
			ID:   call.ID,
			Type: call.Type,
			FunctionCall: &llms.FunctionCall{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		})
	}
	return tools
}

func looksLikeXMLTool(content string) bool {
	return strings.Contains(content, "<function=") || strings.Contains(content, "tool<")
}
