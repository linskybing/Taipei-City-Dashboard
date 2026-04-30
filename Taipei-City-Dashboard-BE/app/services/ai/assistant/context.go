package assistant

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

func ApplyContext(messages []llms.MessageContent, ctx RequestContext) []llms.MessageContent {
	instruction := buildSystemInstruction(ctx)
	enriched := make([]llms.MessageContent, 0, len(messages)+1)
	merged := false

	for _, msg := range messages {
		if msg.Role == llms.ChatMessageTypeSystem && !merged {
			enriched = append(enriched, mergeSystemMessage(msg, instruction))
			merged = true
			continue
		}
		enriched = append(enriched, msg)
	}

	if merged {
		return enriched
	}
	return append([]llms.MessageContent{{
		Role:  llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{llms.TextContent{Text: instruction}},
	}}, enriched...)
}

func ToolOptions() []llms.CallOption {
	return []llms.CallOption{
		llms.WithTools(ToolDefinitions()),
		llms.WithToolChoice("auto"),
	}
}

func ThemeLabel(theme string) string {
	labels := map[string]string{
		"auto":        "自動判斷",
		"commuting":   "智慧通勤",
		"disaster":    "韌性防災",
		"environment": "永續環境",
		"health":      "食安健康",
		"labor":       "勞動福祉",
		"culture":     "文化共融",
	}
	if label, ok := labels[theme]; ok {
		return label
	}
	return labels[DefaultTheme]
}

func buildSystemInstruction(ctx RequestContext) string {
	return fmt.Sprintf(`你是臺北城市儀表板 AI 決策助理。情境：theme=%s(%s), city=%s, audience=%s, dashboard=%s。
用繁中回答，固定分「重點判讀」「可行建議」「資料信心」。
先用 tools 查組件/來源/儀表板；搜尋或推薦組件用 search_components，單一候選用 get_component_snapshot，城市比較用 compare_city_components。
統計工具在要求分析、判讀、資料品質、描述統計、趨勢、季節、異常、檢定、預測，或指定 component_id 時使用；search_components 有候選時，用最高相關 component_id 做描述統計與趨勢初判；完整統計診斷且有 component_id 時可同輪用全部統計工具。
不得編造資料、SQL、即時狀態、模型細節或因果；工具 degraded/unavailable 或資料不足時保守說明限制。
不得透露金鑰或後端設定，不直接串接外部 API；新資料需先匯入、驗證並納入儀表板資料庫。`,
		ctx.Theme,
		ThemeLabel(ctx.Theme),
		ctx.City,
		ctx.Audience,
		emptyAsNone(ctx.DashboardIndex),
	)
}

func mergeSystemMessage(msg llms.MessageContent, instruction string) llms.MessageContent {
	parts := make([]llms.ContentPart, 0, len(msg.Parts)+1)
	merged := false
	for _, part := range msg.Parts {
		text, ok := part.(llms.TextContent)
		if ok && !merged {
			parts = append(parts, llms.TextContent{Text: text.Text + "\n\n" + instruction})
			merged = true
			continue
		}
		parts = append(parts, part)
	}
	if !merged {
		parts = append(parts, llms.TextContent{Text: instruction})
	}
	return llms.MessageContent{Role: msg.Role, Parts: parts}
}

func emptyAsNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "未指定"
	}
	return value
}
