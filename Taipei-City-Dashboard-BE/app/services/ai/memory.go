package ai

import (
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/logs"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

const (
	maxMemoryTurns         = 3
	maxMemoryQuestionRunes = 120
	maxMemoryAnswerRunes   = 180
	maxMemoryBlockRunes    = 900
	memoryEllipsis         = "..."
)

func (s *aiSession) injectMemory() {
	if !shouldLoadMemory(s.req) {
		return
	}
	chatLogs, err := models.GetRecentAIChatLogs(s.req.SessionID, s.req.UserID, maxMemoryTurns)
	if err != nil {
		logs.FError("AI memory load error: %v", err)
		return
	}
	memory := buildSessionMemory(chatLogs)
	if memory == "" {
		return
	}
	s.currentMessages = insertMemoryMessage(s.currentMessages, memory)
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

func buildSessionMemory(chatLogs []models.AIChatLog) string {
	var builder strings.Builder
	count := 0
	for _, chatLog := range chatLogs {
		question := trimMemoryText(chatLog.Question, maxMemoryQuestionRunes)
		answer := trimMemoryText(chatLog.Answer, maxMemoryAnswerRunes)
		if question == "" && answer == "" {
			continue
		}
		count++
		if count == 1 {
			builder.WriteString("對話記憶（最多3輪；若與目前工具結果衝突，以工具結果為準）：")
		}
		builder.WriteString(fmt.Sprintf("\n%d. Q:%s A:%s", count, question, answer))
	}
	if count == 0 {
		return ""
	}
	return trimMemoryText(builder.String(), maxMemoryBlockRunes)
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
