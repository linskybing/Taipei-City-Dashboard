package ai

import (
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/app/services/ai/assistant"
	"TaipeiCityDashboardBE/global"
	"TaipeiCityDashboardBE/logs"
	"context"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms"
)

const maxAllowedToolLoops = 10

type aiSession struct {
	req             AIChatRequest
	options         []llms.CallOption
	callOpts        llms.CallOptions
	currentMessages []llms.MessageContent
	totalInput      int
	totalOutput     int
	toolUsed        bool
	executedTools   []string
	toolResults     []string
	toolEvents      []assistant.ToolEvent
	allowedTools    map[string]bool
	lastResp        *llms.ContentResponse
	lastErr         error
	startTime       time.Time
}

func (s *aiSession) run(ctx context.Context) (*models.AIChatLog, error) {
	s.executedTools = make([]string, 0)
	s.toolResults = make([]string, 0)
	s.toolEvents = make([]assistant.ToolEvent, 0)
	maxLoops := configuredToolLoops()
	for i := 0; i < maxLoops; i++ {
		s.sendHeartbeat(ctx)
		if err := s.generate(ctx); err != nil {
			break
		}
		toolCalls := s.extractToolCalls()
		if len(toolCalls) == 0 {
			break
		}
		s.toolUsed = true
		logs.FInfo("Loop %d: Processing %d tool calls", i, len(toolCalls))
		if err := s.executeTools(ctx, toolCalls, i+1); err != nil {
			break
		}
		if i == maxLoops-1 {
			s.lastErr = fmt.Errorf("tool call loop limit exceeded after %d loops", maxLoops)
			s.recordToolEvent("tool_loop", "error", i+1, 0, s.lastErr)
			break
		}
	}
	return s.finalize()
}

func (s *aiSession) sendHeartbeat(ctx context.Context) {
	if s.callOpts.StreamingFunc != nil {
		s.callOpts.StreamingFunc(ctx, []byte(": heartbeat\n\n"))
	}
}

func (s *aiSession) generate(ctx context.Context) error {
	maxRetry := global.TWCC.MaxRetry
	if s.callOpts.StreamingFunc != nil {
		maxRetry = 0
	}
	for i := 0; i <= maxRetry; i++ {
		s.lastResp, s.lastErr = twccModel.GenerateContent(ctx, s.currentMessages, s.options...)
		if s.lastErr == nil {
			s.updateTokens()
			return nil
		}
		logs.FError("Attempt %d failed: %v", i+1, s.lastErr)
		if i < maxRetry {
			time.Sleep(500 * time.Millisecond)
		}
	}
	return s.lastErr
}

func (s *aiSession) extractToolCalls() []llms.ToolCall {
	if s.lastResp == nil || len(s.lastResp.Choices) == 0 {
		return nil
	}
	tc, _ := s.lastResp.Choices[0].GenerationInfo["tool_calls"].([]llms.ToolCall)
	return tc
}

func (s *aiSession) updateTokens() {
	if s.lastResp == nil || len(s.lastResp.Choices) == 0 {
		return
	}
	if usage, ok := s.lastResp.Choices[0].GenerationInfo["usage"].(map[string]interface{}); ok {
		s.totalInput += parseUsageInt(usage["input_tokens"])
		s.totalOutput += parseUsageInt(usage["output_tokens"])
	}
}

func (s *aiSession) executeTools(ctx context.Context, toolCalls []llms.ToolCall, loop int) error {
	choice := s.lastResp.Choices[0]
	s.currentMessages = append(s.currentMessages, llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: append([]llms.ContentPart{llms.TextContent{Text: choice.Content}}, toolsToParts(toolCalls)...),
	})
	for _, tc := range toolCalls {
		name, args := toolCallPayload(tc)
		auditName := auditToolName(name)
		s.executedTools = append(s.executedTools, auditName)
		start := time.Now()
		result, err := s.executeAllowedTool(ctx, name, args)
		latencyMS := int(time.Since(start).Milliseconds())
		if err != nil {
			s.recordToolEvent(auditName, "error", loop, latencyMS, err)
			result = buildToolFallback(auditName, err)
			logs.FError("Tool Error: %v", err)
		} else {
			s.recordToolEvent(auditName, "success", loop, latencyMS, nil)
		}
		s.toolResults = append(s.toolResults, result)
		s.currentMessages = append(s.currentMessages, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: tc.ID,
				Name:       auditName,
				Content:    result,
			}},
		})
	}
	return nil
}

func (s *aiSession) injectInstructions() {
	toolNames := toolNames(s.callOpts.Tools)
	instruction := fmt.Sprintf(
		"\nSystem Instruction:\n1. Use ONLY: [%s].\n2. NEVER nest tool calls \n3. Arguments MUST be literal values (strings, integers, etc.), never function calls \n4. For dependent tasks, call tools sequentially in separate turns.\n5. If stuck, respond with text.",
		toolNames,
	)
	s.currentMessages = make([]llms.MessageContent, 0, len(s.req.Messages)+1)
	merged := false
	for _, msg := range s.req.Messages {
		if msg.Role == llms.ChatMessageTypeSystem && !merged {
			s.currentMessages = append(s.currentMessages, mergeSystemMsg(msg, instruction))
			merged = true
			continue
		}
		s.currentMessages = append(s.currentMessages, msg)
	}
	if !merged {
		s.currentMessages = append([]llms.MessageContent{{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Instruction: Use tools: [" + toolNames + "]."}},
		}}, s.currentMessages...)
	}
}
