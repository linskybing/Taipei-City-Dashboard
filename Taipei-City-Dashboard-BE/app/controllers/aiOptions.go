package controllers

import "github.com/tmc/langchaingo/llms"

// ToGenerationOptions extracts and maps AI generation options without client tools.
func (input *AIChatInput) ToGenerationOptions() []llms.CallOption {
	options := make([]llms.CallOption, 0)
	params := make(map[string]interface{})

	if input.MaxNewTokens != nil {
		options = append(options, llms.WithMaxTokens(*input.MaxNewTokens))
		params["max_new_tokens"] = *input.MaxNewTokens
	}
	if input.Temperature != nil {
		options = append(options, llms.WithTemperature(*input.Temperature))
		params["temperature"] = *input.Temperature
	}
	if input.TopP != nil {
		options = append(options, llms.WithTopP(*input.TopP))
		params["top_p"] = *input.TopP
	}
	if input.TopK != nil {
		options = append(options, llms.WithTopK(*input.TopK))
		params["top_k"] = *input.TopK
	}
	if input.FrequencePenalty != nil {
		options = append(options, llms.WithRepetitionPenalty(*input.FrequencePenalty))
		params["frequence_penalty"] = *input.FrequencePenalty
	}
	if len(input.StopSequences) > 0 {
		options = append(options, llms.WithStopWords(input.StopSequences))
		params["stop_sequences"] = input.StopSequences
	}
	if input.Seed != nil {
		params["seed"] = *input.Seed
	}
	if len(params) > 0 {
		options = append(options, llms.WithMetadata(params))
	}
	return options
}
