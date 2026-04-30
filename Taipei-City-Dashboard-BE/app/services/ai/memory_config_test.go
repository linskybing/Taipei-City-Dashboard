package ai

import (
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/global"
	"context"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestConfiguredMemoryClampsRuntimeValues(t *testing.T) {
	originalTurns := global.TWCC.MemoryTurns
	originalBlock := global.TWCC.MemoryBlockRunes
	defer func() {
		global.TWCC.MemoryTurns = originalTurns
		global.TWCC.MemoryBlockRunes = originalBlock
	}()

	global.TWCC.MemoryTurns = 0
	global.TWCC.MemoryBlockRunes = 100
	cfg := configuredMemory()
	if cfg.maxTurns != 4 || cfg.blockRunes != 900 {
		t.Fatalf("configuredMemory low clamp = %#v, want turns=4 block=900", cfg)
	}

	global.TWCC.MemoryTurns = 999
	global.TWCC.MemoryBlockRunes = 99999
	cfg = configuredMemory()
	if cfg.maxTurns != 12 || cfg.blockRunes != 2400 {
		t.Fatalf("configuredMemory high clamp = %#v, want turns=12 block=2400", cfg)
	}
}

func TestMetadataParamsWritesZeroMemoryStats(t *testing.T) {
	session := &aiSession{
		req: AIChatRequest{Params: map[string]interface{}{"theme": "health"}},
		memoryStats: memoryStats{
			Source:   memorySourceAIChatLog,
			MaxTurns: 8,
		},
	}
	params := session.metadataParams()
	if params["theme"] != "health" {
		t.Fatalf("metadata params did not preserve base params: %#v", params)
	}
	if params["memory_loaded"] != false || params["memory_turns_loaded"] != 0 {
		t.Fatalf("metadata params should record zero memory stats: %#v", params)
	}
	if params["memory_source"] != memorySourceAIChatLog || params["memory_max_turns"] != 8 {
		t.Fatalf("metadata params missing memory source/max turns: %#v", params)
	}
}

func TestInjectMemorySkipsLongCurrentRequest(t *testing.T) {
	originalLoader := loadRecentAIChatLogs
	defer func() { loadRecentAIChatLogs = originalLoader }()
	loadRecentAIChatLogs = func(ctx context.Context, sessionID, userID string, limit int) ([]models.AIChatLog, error) {
		t.Fatal("memory loader should not run for very long current requests")
		return nil, nil
	}
	session := &aiSession{req: AIChatRequest{
		SessionID: "session_1",
		UserID:    "7",
		Messages: []llms.MessageContent{{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: strings.Repeat("長", maxCurrentRequestRunesWithMemory+1)}},
		}},
	}}
	session.injectMemory(context.Background())
	if session.memoryStats.Loaded || !session.memoryStats.Truncated {
		t.Fatalf("long request should skip loaded memory and mark truncation: %#v", session.memoryStats)
	}
}
