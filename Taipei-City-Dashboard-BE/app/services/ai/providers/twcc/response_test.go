package twcc

import (
	"context"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestStandardResponseExtractsXMLToolCalls(t *testing.T) {
	model := &TWCC{ModelName: "llama3.3-ffm-70b-16k-chat"}
	body := strings.NewReader(
		`{"generated_text":"說明 <function=recommend_actions>{\"theme\":\"health\"}</function> 結束","prompt_tokens":1,"generated_tokens":2,"total_tokens":3}`,
	)

	resp, err := model.handleStandardResponse(body)
	if err != nil {
		t.Fatalf("handleStandardResponse returned error: %v", err)
	}
	choice := resp.Choices[0]
	if strings.Contains(choice.Content, "<function=") {
		t.Fatalf("content still contains raw tool markup: %q", choice.Content)
	}
	assertToolCall(t, choice.GenerationInfo["tool_calls"], "recommend_actions")
}

func TestStreamingResponseSuppressesXMLToolMarkup(t *testing.T) {
	ctx := context.Background()
	var streamed strings.Builder
	proc := &streamProcessor{
		toolCallsMap: make(map[int]*TWCCToolCall),
		streamingFunc: func(ctx context.Context, chunk []byte) error {
			streamed.Write(chunk)
			return nil
		},
	}

	proc.processLine(ctx,
		"data: {\"generated_text\":\"<function=recommend_actions>{\\\"theme\\\":\\\"health\\\"}</function>\",\"prompt_tokens\":1,\"generated_tokens\":2,\"total_tokens\":3}\n",
	)
	proc.finalize(ctx)
	resp := proc.toContentResponse("llama3.3-ffm-70b-16k-chat")

	if strings.Contains(streamed.String(), "<function=") {
		t.Fatalf("streamed raw tool markup: %q", streamed.String())
	}
	if strings.Contains(resp.Choices[0].Content, "<function=") {
		t.Fatalf("response content still contains tool markup: %q", resp.Choices[0].Content)
	}
	assertToolCall(t, resp.Choices[0].GenerationInfo["tool_calls"], "recommend_actions")
}

func assertToolCall(t *testing.T, raw interface{}, name string) {
	t.Helper()
	calls, ok := raw.([]llms.ToolCall)
	if !ok || len(calls) != 1 {
		t.Fatalf("tool_calls = %#v, want one llms.ToolCall", raw)
	}
	if calls[0].FunctionCall == nil || calls[0].FunctionCall.Name != name {
		t.Fatalf("tool call = %#v, want %q", calls[0], name)
	}
}
