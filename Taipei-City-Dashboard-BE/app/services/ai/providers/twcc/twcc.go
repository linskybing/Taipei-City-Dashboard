package twcc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"TaipeiCityDashboardBE/logs"
	"github.com/tmc/langchaingo/llms"
)

type TWCC struct {
	APIKey      string
	BaseURL     string
	ModelName   string
	HTTPClient  *http.Client
	Temperature float64
	MaxTokens   int
}

var _ llms.Model = (*TWCC)(nil)

func New(apiKey, baseURL, model string, timeout int) *TWCC {
	return &TWCC{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		ModelName:   model,
		HTTPClient:  &http.Client{Timeout: time.Duration(timeout) * time.Second},
		Temperature: 0.7,
		MaxTokens:   350,
	}
}

func (m *TWCC) GenerateContent(
	ctx context.Context,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	reqBody := TWCCRequest{
		Model:      m.ModelName,
		Messages:   m.toTWCCMessages(messages),
		Parameters: m.toTWCCParameters(&opts),
		Stream:     opts.StreamingFunc != nil,
		Tools:      m.toTWCCTools(opts.Tools),
	}
	if opts.ToolChoice != nil {
		reqBody.ToolChoice = opts.ToolChoice
	} else if len(reqBody.Tools) > 0 {
		reqBody.ToolChoice = "auto"
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	logs.FInfo("TWCC Outgoing Request: %s", string(jsonData))

	resp, err := m.doRequest(ctx, jsonData, reqBody.Stream)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if reqBody.Stream {
		return m.handleStreamingResponse(ctx, resp.Body, &opts)
	}
	return m.handleStandardResponse(resp.Body)
}

func (m *TWCC) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	msg := llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
	}
	resp, err := m.GenerateContent(ctx, []llms.MessageContent{msg}, options...)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Content, nil
}
