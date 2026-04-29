package controllers

import "github.com/tmc/langchaingo/llms"

// AIChatInput matches the TWCC request schema with dashboard assistant context.
type AIChatInput struct {
	SessionID      string `json:"session"`
	Stream         bool   `json:"stream"`
	Theme          string `json:"theme" binding:"omitempty,oneof=auto commuting disaster environment health labor culture"`
	City           string `json:"city" binding:"omitempty,oneof=taipei metrotaipei"`
	Audience       string `json:"audience" binding:"omitempty,oneof=government public"`
	DashboardIndex string `json:"dashboard_index,omitempty"`
	Messages       []struct {
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
	} `json:"messages" binding:"required,gt=0"`
	MaxNewTokens     *int     `json:"max_new_tokens" binding:"omitempty,gt=0"`
	Temperature      *float64 `json:"temperature" binding:"omitempty,gt=0"`
	TopP             *float64 `json:"top_p" binding:"omitempty,gt=0,lte=1"`
	TopK             *int     `json:"top_k" binding:"omitempty,gte=1,lte=100"`
	FrequencePenalty *float64 `json:"frequence_penalty" binding:"omitempty,gt=0"`
	StopSequences    []string `json:"stop_sequences" binding:"omitempty,max=4"`
	Seed             *int     `json:"seed" binding:"omitempty,gte=0"`
	Tools            []struct {
		Type     string `json:"type" binding:"required,eq=function"`
		Function struct {
			Name        string      `json:"name" binding:"required"`
			Description string      `json:"description,omitempty"`
			Parameters  interface{} `json:"parameters,omitempty"`
		} `json:"function" binding:"required"`
	} `json:"tools,omitempty"`
	ToolChoice interface{} `json:"tool_choice,omitempty"`
}

// ToServiceMessages converts input messages to langchaingo internal format.
func (input *AIChatInput) ToServiceMessages() []llms.MessageContent {
	serviceMsgs := make([]llms.MessageContent, 0, len(input.Messages))
	for _, m := range input.Messages {
		role := llms.ChatMessageTypeHuman
		parts := []llms.ContentPart{llms.TextContent{Text: m.Content}}
		switch m.Role {
		case "assistant":
			role = llms.ChatMessageTypeAI
		case "system":
			continue
		case "tool":
			continue
		}
		serviceMsgs = append(serviceMsgs, llms.MessageContent{Role: role, Parts: parts})
	}
	return serviceMsgs
}
