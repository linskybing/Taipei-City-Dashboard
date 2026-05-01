package assistant

import (
	"fmt"
	"sort"
	"strings"
)

func searchComponentsEnvelope(
	query string,
	related []RelatedComponent,
	notes []string,
	data map[string]interface{},
) toolEnvelope {
	if data == nil {
		data = map[string]interface{}{}
	}
	ranked := rankRelatedComponents(query, related)
	assessment := assessSearchConfidence(query, ranked, isDegradedSearch(data))
	data["candidate_count"] = len(ranked)
	data["recommendation_count"] = len(assessment.recommended)
	data["confidence_band"] = string(assessment.band)
	if _, ok := data["status"]; !ok {
		data["status"] = "ok"
	}
	if dropCount := cityResolutionDropCount(data); dropCount > 0 {
		notes = append(notes, crossCityResolutionNote(dropCount, cityResolutionDropPreviewFromData(data)))
	}
	if assessment.band == searchConfidenceLow && data["status"] == "ok" {
		data["status"] = "low_confidence"
	}
	notes = append(notes, assessment.notes...)
	cards, cardNotes := searchAnalysisCards(query, assessment.analysisCandidates)
	notes = append(notes, cardNotes...)
	return toolEnvelope{
		Tool:              "search_components",
		RelatedComponents: assessment.recommended,
		ConfidenceNotes:   notes,
		Guardrails:        componentSearchGuardrails(),
		AnalysisCards:     cards,
		Data:              data,
	}
}

func rankRelatedComponents(query string, related []RelatedComponent) []RelatedComponent {
	ranked := append([]RelatedComponent(nil), related...)
	tokens := queryTokens(query)
	sort.SliceStable(ranked, func(i, j int) bool {
		return relatedTextScore(query, tokens, ranked[i]) > relatedTextScore(query, tokens, ranked[j])
	})
	return ranked
}

func relatedTextScore(query string, tokens []string, component RelatedComponent) float64 {
	text := strings.Join([]string{component.Index, component.Name, component.QueryType}, " ")
	return component.Score + textScore(query, tokens, text)
}

func cityResolutionDropCount(data map[string]interface{}) int {
	if data == nil {
		return 0
	}
	switch value := data["city_resolution_drop_count"].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func cityResolutionDropPreviewFromData(data map[string]interface{}) []cityResolutionDropPreview {
	if data == nil {
		return nil
	}
	preview, _ := data["city_resolution_drop_preview"].([]cityResolutionDropPreview)
	return preview
}

func crossCityResolutionNote(dropCount int, preview []cityResolutionDropPreview) string {
	if len(preview) == 0 {
		return fmt.Sprintf("另有 %d 個相近候選只在其他城市有對應組件，因目前指定城市找不到同 index 組件已先排除；這不代表沒有相近主題，可改指定另一城市或使用 compare_city_components 再確認。", dropCount)
	}
	labels := make([]string, 0, len(preview))
	for _, item := range preview {
		labels = append(labels, cityResolutionDropLabel(item))
	}
	return fmt.Sprintf("另有 %d 個相近候選只在其他城市有對應組件，例如 %s；因目前指定城市找不到同 index 組件已先排除。這不代表沒有相近主題，可改指定另一城市或使用 compare_city_components 再確認。", dropCount, strings.Join(labels, "、"))
	}

func cityResolutionDropLabel(item cityResolutionDropPreview) string {
	if item.Index != "" && item.Name != "" {
		return fmt.Sprintf("%s（%s）", item.Index, item.Name)
	}
	if item.Index != "" {
		return item.Index
	}
	if item.Name != "" {
		return item.Name
	}
	return "未命名候選"
}

func searchAnalysisCards(query string, related []RelatedComponent) ([]AnalysisCard, []string) {
	if !requestsStatAnalysis(query) || len(related) == 0 {
		return nil, nil
	}
	notes := make([]string, 0, 1)
	for _, component := range related {
		cards, err := buildComponentAnalysisCards(component)
		if err != nil {
			notes = append(notes, fmt.Sprintf("component %d 暫無法產生統計卡：%s。", component.ID, err.Error()))
			continue
		}
		notes = append(notes, "已根據最高相關 component 產生描述統計與趨勢初判。")
		return cards, notes
	}
	return nil, notes
}

func buildComponentAnalysisCards(component RelatedComponent) ([]AnalysisCard, error) {
	input := statToolInput{
		ComponentID: int(component.ID),
		City:        component.City,
		Options:     map[string]interface{}{"max_results": 5},
	}
	ds, err := loadStatDataset(input)
	if err != nil {
		return nil, err
	}
	points, err := selectStatPoints(ds, "")
	if err != nil {
		return nil, err
	}
	cards := make([]AnalysisCard, 0, 2)
	if card, err := buildDescriptiveCard(ds, points, input); err == nil {
		cards = append(cards, card)
	}
	if card, err := buildTrendCard(ds, points, input); err == nil {
		cards = append(cards, card)
	}
	if len(cards) == 0 {
		return nil, fmt.Errorf("no supported statistical card")
	}
	return cards, nil
}

func requestsStatAnalysis(query string) bool {
	query = strings.TrimSpace(query)
	for _, token := range []string{
		"分析", "統計", "判讀", "趨勢", "變化", "異常", "預測",
		"資料品質", "描述", "平均", "中位數", "高齡就業人口",
	} {
		if strings.Contains(query, token) {
			return true
		}
	}
	return false
}
