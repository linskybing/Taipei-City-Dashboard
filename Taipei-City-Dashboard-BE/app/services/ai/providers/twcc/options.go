package twcc

import "github.com/tmc/langchaingo/llms"

func (m *TWCC) toTWCCParameters(opts *llms.CallOptions) TWCCParameters {
	params := TWCCParameters{Stream: opts.StreamingFunc != nil}
	meta := opts.Metadata

	if value, ok := meta["max_new_tokens"].(int); ok {
		params.MaxNewTokens = &value
	}
	if value, ok := meta["temperature"].(float64); ok {
		params.Temperature = &value
	}
	if value, ok := meta["top_p"].(float64); ok {
		params.TopP = &value
	}
	if value, ok := meta["top_k"].(int); ok {
		params.TopK = &value
	}
	if value, ok := meta["frequence_penalty"].(float64); ok {
		params.FrequencePenalty = &value
	}
	if value, ok := meta["stop_sequences"].([]string); ok {
		params.StopSequences = value
	}
	if value, ok := meta["seed"].(int); ok {
		params.Seed = &value
	}
	return params
}

func (m *TWCC) toTWCCTools(tools []llms.Tool) []TWCCTool {
	if len(tools) == 0 {
		return nil
	}
	twccTools := make([]TWCCTool, 0, len(tools))
	for _, tool := range tools {
		if tool.Function == nil {
			continue
		}
		params := tool.Function.Parameters
		if params == nil {
			params = map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			}
		}
		twccTools = append(twccTools, TWCCTool{
			Type: tool.Type,
			Function: TWCCToolFunction{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  params,
			},
		})
	}
	return twccTools
}
