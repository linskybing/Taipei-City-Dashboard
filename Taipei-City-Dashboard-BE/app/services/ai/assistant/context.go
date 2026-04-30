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
	playbook := playbookFor(ctx.Theme)
	return fmt.Sprintf(`你是臺北城市儀表板的六主題 AI 決策助理。
使用情境：
- 主題：%s (%s)
- 城市範圍：%s
- 受眾：%s
- 儀表板索引：%s

決策模板：
- 決策問題：%s
- 地圖圖層：%s
- 趨勢圖：%s
- 比較/排行圖：%s
- 優先訊號：%s
- 行動步驟：%s
- 保守邊界：%s

規則：
1. 必須優先使用後端提供的 tools 取得組件、資料來源與儀表板脈絡。
2. 不得編造資料、SQL、模型、來源或即時狀態；資料不足時要明確說明限制。
3. 回答使用繁體中文，給出可執行的市政/民眾決策建議。
4. 若涉及城市比較，使用 compare_city_components。
	5. 若問題很廣，先使用 search_components，再視需要使用 get_component_snapshot。
	6. 使用者說「查看組件」「找組件」「推薦組件」「有哪些儀表板/組件」或「想看某議題資料」時，視為組件搜尋意圖，優先呼叫 search_components；需要核對單一候選的描述或來源時再呼叫 get_component_snapshot。
	7. 不直接呼叫模型供應商，不透露金鑰、模型細節或後端設定。
	8. 統計工具只在使用者明確要求資料品質、描述統計、趨勢、季節、異常、檢定、預測，或使用者指定 component_id 時呼叫；一般儀表板規劃、模板、建議題不得主動呼叫統計工具。
	9. 若工具回傳 degraded 或候選脈絡，應停止追加高風險工具呼叫，直接保守整理已取得訊號與限制。
	10. 除非工具明確回傳因果設計，否則不得使用「造成」「導致」「政策效果」等因果語彙。
	11. 最後輸出要包含「重點判讀」、「可行建議」、「資料信心」三段。
%s`,
		ctx.Theme,
		ThemeLabel(ctx.Theme),
		ctx.City,
		ctx.Audience,
		emptyAsNone(ctx.DashboardIndex),
		playbook.Question,
		playbook.MapLayer,
		playbook.TrendChart,
		playbook.RankChart,
		strings.Join(playbook.Signals, "、"),
		strings.Join(playbook.Steps, " → "),
		strings.Join(playbook.Guardrails, "；"),
		statRoutingInstruction(),
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
