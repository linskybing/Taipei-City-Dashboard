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
用繁中，分「重點判讀」「建議」「資料信心」。
先用 tools 查組件/來源；推薦組件/搜尋用 search_components，單候選查 get_component_snapshot，跨城查 compare_city_components。
search_components 低信心/degraded 時，只能說明限制與下一步，不得直接推薦 component。
分析、趨勢、異常、預測或指定 component_id 時用統計工具；僅高信心 search_components 候選做統計初判；完整統計診斷且有 component_id 時可同輪用全部統計工具。
不得編造資料、SQL、現況、模型細節或因果；工具 unavailable/degraded 或資料不足時保守說明限制。
不透露金鑰/後端設定，不直接串接外部 API；新資料需先匯入、驗證並納入儀表板資料庫。`,
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
