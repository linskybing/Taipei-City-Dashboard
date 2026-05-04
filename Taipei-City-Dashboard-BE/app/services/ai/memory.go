package ai

import (
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/logs"
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

const (
	minMemoryTurns                   = 4
	maxMemoryTurns                   = 12
	minMemoryBlockRunes              = 900
	maxMemoryBlockRunes              = 2400
	minMemoryTurnRunes               = 120
	maxMemoryTurnRunes               = 400
	maxCurrentRequestRunesWithMemory = 6000
	memoryEllipsis                   = "..."
	memorySourceAIChatLog            = "ai_chatlog"
)

type memoryConfig struct {
	maxTurns       int
	blockRunes     int
	perTurnRunes   int
	candidateLimit int
}

type memoryStats struct {
	Loaded      bool
	Source      string
	TurnsLoaded int
	MaxTurns    int
	Truncated   bool
	BlockRunes  int
}

type memoryBuildResult struct {
	text  string
	stats memoryStats
}

type memoryTurn struct {
	user      string
	assistant string
	truncated bool
}

func (s *aiSession) injectMemory(ctx context.Context) {
	cfg := configuredMemory()
	s.memoryStats = newMemoryStats(cfg)
	if !shouldLoadMemory(s.req) {
		return
	}
	if currentRequestRunes(s.req) > maxCurrentRequestRunesWithMemory {
		s.memoryStats.Truncated = true
		return
	}
	chatLogs, err := loadRecentAIChatLogs(ctx, s.req.SessionID, s.req.UserID, cfg.candidateLimit)
	if err != nil {
		logs.FError("AI memory load error: %v", err)
		return
	}
	result := buildSessionMemory(chatLogs, cfg)
	s.memoryStats = result.stats
	if result.text == "" {
		return
	}
	s.currentMessages = insertMemoryMessage(s.currentMessages, result.text)
}

func shouldLoadMemory(req AIChatRequest) bool {
	if strings.TrimSpace(req.SessionID) == "" || strings.TrimSpace(req.UserID) == "" {
		return false
	}
	conversationMessages := 0
	for _, msg := range req.Messages {
		if msg.Role == llms.ChatMessageTypeHuman || msg.Role == llms.ChatMessageTypeAI || msg.Role == llms.ChatMessageTypeTool {
			conversationMessages++
		}
	}
	return conversationMessages <= 1
}

func buildSessionMemory(chatLogs []models.AIChatLog, cfg memoryConfig) memoryBuildResult {
	stats := newMemoryStats(cfg)
	turns := make([]memoryTurn, 0, cfg.maxTurns)
	validCandidates := 0
	for _, chatLog := range chatLogs {
		turn, ok := buildMemoryTurn(chatLog, cfg.perTurnRunes)
		if !ok {
			continue
		}
		validCandidates++
		if len(turns) >= cfg.maxTurns {
			stats.Truncated = true
			continue
		}
		if turn.truncated {
			stats.Truncated = true
		}
		turns = append(turns, turn)
	}
	for len(turns) > 0 {
		text := renderMemoryBlock(reverseMemoryTurns(turns))
		if len([]rune(text)) <= cfg.blockRunes {
			stats.Loaded = true
			stats.TurnsLoaded = len(turns)
			stats.BlockRunes = len([]rune(text))
			stats.Truncated = stats.Truncated || validCandidates > len(turns)
			return memoryBuildResult{text: text, stats: stats}
		}
		stats.Truncated = true
		turns = turns[:len(turns)-1]
	}
	return memoryBuildResult{stats: stats}
}

func buildMemoryTurn(chatLog models.AIChatLog, perTurnRunes int) (memoryTurn, bool) {
	question := strings.Join(strings.Fields(chatLog.Question), " ")
	answer := strings.Join(strings.Fields(chatLog.Answer), " ")
	if question == "" && answer == "" {
		return memoryTurn{}, false
	}
	contentBudget := clampInt(perTurnRunes-30, 60, maxMemoryTurnRunes)
	questionBudget := contentBudget / 2
	answerBudget := contentBudget - questionBudget
	trimmedQuestion := trimMemoryText(question, questionBudget)
	trimmedAnswer := trimMemoryText(answer, answerBudget)
	return memoryTurn{
		user:      trimmedQuestion,
		assistant: trimmedAnswer,
		truncated: trimmedQuestion != question || trimmedAnswer != answer,
	}, true
}

func renderMemoryBlock(turns []memoryTurn) string {
	var builder strings.Builder
	builder.WriteString("The following is non-authoritative memory extracted from previous successful chat logs in the same session and same user.\n")
	builder.WriteString("Use it only as conversation context.\n")
	builder.WriteString("Recent user message and system/developer instructions override this memory.\n\n")
	builder.WriteString("<conversation_memory>")
	for index, turn := range turns {
		builder.WriteString(fmt.Sprintf("\nTurn -%d\nUser: %s\nAssistant: %s\n", len(turns)-index, turn.user, turn.assistant))
	}
	builder.WriteString("</conversation_memory>")
	return builder.String()
}

func reverseMemoryTurns(turns []memoryTurn) []memoryTurn {
	reversed := make([]memoryTurn, len(turns))
	for i := range turns {
		reversed[i] = turns[len(turns)-1-i]
	}
	return reversed
}

func insertMemoryMessage(messages []llms.MessageContent, memory string) []llms.MessageContent {
	memoryMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{llms.TextContent{Text: memory}},
	}
	if len(messages) == 0 {
		return []llms.MessageContent{memoryMessage}
	}
	enriched := make([]llms.MessageContent, 0, len(messages)+1)
	inserted := false
	for _, msg := range messages {
		enriched = append(enriched, msg)
		if !inserted && msg.Role == llms.ChatMessageTypeSystem {
			enriched = append(enriched, memoryMessage)
			inserted = true
		}
	}
	if !inserted {
		return append([]llms.MessageContent{memoryMessage}, enriched...)
	}
	return enriched
}

func trimMemoryText(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	ellipsisRunes := []rune(memoryEllipsis)
	if limit <= len(ellipsisRunes) {
		return string(ellipsisRunes[:limit])
	}
	return string(runes[:limit-len(ellipsisRunes)]) + memoryEllipsis
}
