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
	related = rankRelatedComponents(query, related)
	cards, cardNotes := searchAnalysisCards(query, related)
	notes = append(notes, cardNotes...)
	return toolEnvelope{
		Tool:              "search_components",
		RelatedComponents: related,
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
