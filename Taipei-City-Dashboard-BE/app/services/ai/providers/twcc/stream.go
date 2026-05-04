package twcc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

func (m *TWCC) handleStreamingResponse(
	ctx context.Context,
	body io.Reader,
	opts *llms.CallOptions,
) (*llms.ContentResponse, error) {
	reader := bufio.NewReader(body)
	proc := &streamProcessor{
		toolCallsMap:  make(map[int]*TWCCToolCall),
		streamingFunc: opts.StreamingFunc,
	}

	for {
		line, err := reader.ReadString('\n')
		if line != "" && proc.processLine(ctx, line) {
			break
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading stream: %v", err)
		}
	}

	proc.finalize(ctx)
	return proc.toContentResponse(m.ModelName), nil
}

type streamProcessor struct {
	isToolCalling      bool
	detectionConfirmed bool
	lineBuffer         []string
	contentBuffer      strings.Builder
	toolCallsMap       map[int]*TWCCToolCall
	fullContent        strings.Builder
	lastUsage          *TWCCStreamResponse
	streamingFunc      func(context.Context, []byte) error
}

func (p *streamProcessor) processLine(ctx context.Context, line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "data:") {
		p.handleControlLine(ctx, line)
		return false
	}

	jsonData := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
	if jsonData == "[DONE]" {
		p.handleDone(ctx, line)
		return true
	}

	var chunk TWCCStreamResponse
	if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
		return false
	}
	p.detectType(&chunk)
	p.processChunk(ctx, &chunk, line)
	return false
}

func (p *streamProcessor) handleControlLine(ctx context.Context, line string) {
	if !p.detectionConfirmed {
		p.lineBuffer = append(p.lineBuffer, line)
		return
	}
	if !p.isToolCalling && p.streamingFunc != nil {
		p.streamingFunc(ctx, []byte(line))
	}
}

func (p *streamProcessor) handleDone(ctx context.Context, line string) {
	if !p.detectionConfirmed {
		p.isToolCalling = false
		p.detectionConfirmed = true
		p.flushBuffer(ctx)
	}
	if !p.isToolCalling && p.streamingFunc != nil {
		p.streamingFunc(ctx, []byte(line))
	}
}

func (p *streamProcessor) detectType(chunk *TWCCStreamResponse) {
	if p.detectionConfirmed {
		return
	}
	if len(chunk.ToolCalls) > 0 || hasDeltaToolCalls(chunk) {
		p.isToolCalling = true
		p.detectionConfirmed = true
		return
	}

	text := streamText(chunk)
	p.contentBuffer.WriteString(text)
	combined := p.contentBuffer.String()
	if looksLikeXMLTool(combined) {
		p.isToolCalling = true
		p.detectionConfirmed = true
		return
	}
	if len(combined) > 64 {
		p.isToolCalling = false
		p.detectionConfirmed = true
	}
}

func (p *streamProcessor) processChunk(ctx context.Context, chunk *TWCCStreamResponse, rawLine string) {
	p.fullContent.WriteString(streamText(chunk))
	if chunk.PromptTokens > 0 || chunk.Usage != nil {
		p.lastUsage = chunk
	}
	if !p.detectionConfirmed {
		p.lineBuffer = append(p.lineBuffer, rawLine)
		return
	}
	if p.isToolCalling {
		p.accumulateTools(chunk)
		return
	}
	p.flushBuffer(ctx)
	if p.streamingFunc != nil {
		p.streamingFunc(ctx, []byte(rawLine))
	}
}

func (p *streamProcessor) accumulateTools(chunk *TWCCStreamResponse) {
	source := chunk.ToolCalls
	if len(chunk.Choices) > 0 && len(chunk.Choices[0].Delta.ToolCalls) > 0 {
		source = chunk.Choices[0].Delta.ToolCalls
	}
	for _, tc := range source {
		idx := 0
		if tc.Index != nil {
			idx = *tc.Index
		}
		if _, exists := p.toolCallsMap[idx]; !exists {
			p.toolCallsMap[idx] = &TWCCToolCall{ID: tc.ID, Type: tc.Type}
		}
		p.toolCallsMap[idx].Function.Arguments += cleanXML(tc.Function.Arguments)
		if tc.ID != "" {
			p.toolCallsMap[idx].ID = tc.ID
		}
		if tc.Type != "" {
			p.toolCallsMap[idx].Type = tc.Type
		}
		if tc.Function.Name != "" {
			p.toolCallsMap[idx].Function.Name = tc.Function.Name
		}
	}
}

func (p *streamProcessor) flushBuffer(ctx context.Context) {
	if len(p.lineBuffer) == 0 || p.streamingFunc == nil {
		p.lineBuffer = nil
		return
	}
	for _, line := range p.lineBuffer {
		p.streamingFunc(ctx, []byte(line))
	}
	p.lineBuffer = nil
}
