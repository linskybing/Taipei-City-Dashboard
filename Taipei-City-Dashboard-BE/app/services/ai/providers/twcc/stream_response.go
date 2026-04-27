package twcc

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

func (p *streamProcessor) finalize(ctx context.Context) {
	if !p.detectionConfirmed {
		p.isToolCalling = false
		p.flushBuffer(ctx)
	}
}

func (p *streamProcessor) toContentResponse(model string) *llms.ContentResponse {
	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{
			Content:        p.fullContent.String(),
			GenerationInfo: map[string]interface{}{"model": model},
		}},
	}
	if p.isToolCalling {
		p.attachToolCalls(resp)
	}
	if p.lastUsage != nil {
		resp.Choices[0].GenerationInfo["usage"] = streamUsage(p.lastUsage)
	}
	return resp
}

func (p *streamProcessor) attachToolCalls(resp *llms.ContentResponse) {
	tools := make([]llms.ToolCall, 0, len(p.toolCallsMap))
	for _, tc := range p.toolCallsMap {
		tools = append(tools, llms.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			FunctionCall: &llms.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	if len(tools) == 0 {
		extracted, cleaned := extractXMLToolCalls(p.fullContent.String())
		if len(extracted) > 0 {
			tools = extracted
			resp.Choices[0].Content = cleaned
		}
	}
	resp.Choices[0].GenerationInfo["tool_calls"] = tools
}

func streamUsage(resp *TWCCStreamResponse) map[string]interface{} {
	inputTokens := resp.PromptTokens
	outputTokens := resp.GeneratedTokens
	totalTokens := resp.TotalTokens
	if resp.Usage != nil {
		if inputTokens == 0 {
			inputTokens = resp.Usage.PromptTokens
		}
		if outputTokens == 0 {
			outputTokens = resp.Usage.GeneratedTokens
		}
		totalTokens = inputTokens + outputTokens
	}
	return map[string]interface{}{
		"input_tokens":  inputTokens,
		"output_tokens": outputTokens,
		"total_tokens":  totalTokens,
	}
}

func hasDeltaToolCalls(chunk *TWCCStreamResponse) bool {
	return len(chunk.Choices) > 0 &&
		(len(chunk.Choices[0].Delta.ToolCalls) > 0 ||
			chunk.Choices[0].FinishReason == "tool_calls")
}

func streamText(chunk *TWCCStreamResponse) string {
	if len(chunk.Choices) > 0 {
		return chunk.Choices[0].Delta.Content
	}
	return chunk.GeneratedText
}
