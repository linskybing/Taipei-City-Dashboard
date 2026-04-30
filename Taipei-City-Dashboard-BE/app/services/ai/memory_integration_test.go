package ai

import (
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/global"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestChatWithTWCCInjectsMemoryAndWritesStats(t *testing.T) {
	originalModel := twccModel
	originalLoader := loadRecentAIChatLogs
	originalWriter := createAIChatLog
	originalTurns := global.TWCC.MemoryTurns
	originalBlock := global.TWCC.MemoryBlockRunes
	defer func() {
		twccModel = originalModel
		loadRecentAIChatLogs = originalLoader
		createAIChatLog = originalWriter
		global.TWCC.MemoryTurns = originalTurns
		global.TWCC.MemoryBlockRunes = originalBlock
	}()

	global.TWCC.MemoryTurns = 8
	global.TWCC.MemoryBlockRunes = 1600
	ctx := context.WithValue(context.Background(), testContextKey{}, "request")
	fake := &fakeMemoryModel{}
	twccModel = fake
	loadRecentAIChatLogs = fakeMemoryLoader(t)

	var savedLog *models.AIChatLog
	createAIChatLog = func(ctx context.Context, log *models.AIChatLog) error {
		if ctx.Value(testContextKey{}) != "request" {
			t.Fatal("chat log writer did not receive request context")
		}
		copyLog := *log
		savedLog = &copyLog
		return nil
	}

	logEntry, err := ChatWithTWCC(ctx, AIChatRequest{
		SessionID: "session_1",
		UserID:    "7",
		IPAddress: "127.0.0.1",
		Messages: []llms.MessageContent{{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "第五題呢？"}},
		}},
		Params: map[string]interface{}{"theme": "health"},
	})
	if err != nil {
		t.Fatalf("ChatWithTWCC returned error: %v", err)
	}
	if savedLog == nil || logEntry.Answer != savedLog.Answer {
		t.Fatalf("expected chat log to be saved, got returned=%#v saved=%#v", logEntry, savedLog)
	}
	memoryText := fake.systemText()
	for _, want := range []string{"<conversation_memory>", "User: 第一題", "User: 第四題"} {
		if !strings.Contains(memoryText, want) {
			t.Fatalf("model messages missing %q: %s", want, memoryText)
		}
	}
	assertMemoryMetadata(t, savedLog.Metadata)
}

func fakeMemoryLoader(t *testing.T) func(context.Context, string, string, int) ([]models.AIChatLog, error) {
	return func(ctx context.Context, sessionID, userID string, limit int) ([]models.AIChatLog, error) {
		if ctx.Value(testContextKey{}) != "request" {
			t.Fatal("memory loader did not receive request context")
		}
		if sessionID != "session_1" || userID != "7" || limit != 16 {
			t.Fatalf("loader args = %q/%q/%d", sessionID, userID, limit)
		}
		return []models.AIChatLog{
			{Question: "第四題", Answer: "第四答"},
			{Question: "第三題", Answer: "第三答"},
			{Question: "第二題", Answer: "第二答"},
			{Question: "第一題", Answer: "第一答"},
		}, nil
	}
}

type testContextKey struct{}

type fakeMemoryModel struct {
	messages []llms.MessageContent
}

func (m *fakeMemoryModel) GenerateContent(
	ctx context.Context,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	m.messages = messages
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{
		Content: "測試回答",
		GenerationInfo: map[string]interface{}{
			"usage": map[string]interface{}{"input_tokens": 10, "output_tokens": 3},
		},
	}}}, nil
}

func (m *fakeMemoryModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "測試回答", nil
}

func (m *fakeMemoryModel) systemText() string {
	var builder strings.Builder
	for _, msg := range m.messages {
		if msg.Role != llms.ChatMessageTypeSystem {
			continue
		}
		builder.WriteString(extractText(msg))
		builder.WriteString("\n")
	}
	return builder.String()
}

func assertMemoryMetadata(t *testing.T, raw string) {
	t.Helper()
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		t.Fatalf("metadata is invalid JSON: %v; raw=%s", err, raw)
	}
	if metadata["theme"] != "health" {
		t.Fatalf("metadata did not preserve existing fields: %#v", metadata)
	}
	if loaded, _ := metadata["memory_loaded"].(bool); !loaded {
		t.Fatalf("memory_loaded = %#v, want true", metadata["memory_loaded"])
	}
	if turns, _ := metadata["memory_turns_loaded"].(float64); turns != 4 {
		t.Fatalf("memory_turns_loaded = %#v, want 4", metadata["memory_turns_loaded"])
	}
	if source, _ := metadata["memory_source"].(string); source != memorySourceAIChatLog {
		t.Fatalf("memory_source = %q, want %q", source, memorySourceAIChatLog)
	}
	if runes, _ := metadata["memory_block_runes"].(float64); runes <= 0 {
		t.Fatalf("memory_block_runes = %#v, want positive", metadata["memory_block_runes"])
	}
	if strings.Contains(raw, "conversation_memory") || strings.Contains(raw, "第一題") {
		t.Fatalf("metadata must not store raw memory text: %s", raw)
	}
}
