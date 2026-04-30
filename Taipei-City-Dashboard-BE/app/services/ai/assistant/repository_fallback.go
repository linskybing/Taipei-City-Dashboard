package assistant

import (
	"TaipeiCityDashboardBE/app/models"
	"sort"
	"strings"
)

type dashboardCandidate struct {
	Index string  `json:"index"`
	Name  string  `json:"name"`
	City  string  `json:"city"`
	Score float64 `json:"score"`
}

func searchComponentsByMetadata(query string, city string, limit int) ([]RelatedComponent, []string, error) {
	components, _, _, err := models.GetAllComponents(city, 0, 0, "", "", "", "", "", "", "")
	if err != nil {
		return nil, nil, err
	}
	tokens := queryTokens(query)
	related := make([]RelatedComponent, 0, len(components))
	for _, component := range components {
		score := componentMetadataScore(query, tokens, component)
		if score > 0 {
			related = append(related, relatedFromComponent(component, score))
		}
	}
	sort.SliceStable(related, func(i, j int) bool {
		return related[i].Score > related[j].Score
	})
	notes := []string{"語意索引暫不可用，已改用既有 dashboard metadata 關鍵字比對。"}
	if len(related) == 0 {
		notes = append(notes, "未找到關鍵字相符組件，提供同城市既有組件作為保守脈絡。")
		for _, component := range components {
			related = append(related, relatedFromComponent(component, 0))
			if len(related) >= limit {
				break
			}
		}
	}
	if len(related) > limit {
		related = related[:limit]
	}
	return related, notes, nil
}

func buildSearchComponentsFallback(query string, city string, limit int) (string, error) {
	related, notes, err := searchComponentsByMetadata(query, city, limit)
	if err != nil {
		return unavailableTool("search_components", err)
	}
	return marshalTool(searchComponentsEnvelope(query, related, notes, map[string]interface{}{
		"fallback":     "db_metadata",
		"query":        query,
		"result_count": len(related),
		"status":       "degraded",
	}))
}

func buildDashboardContextFallback(req RequestContext) (string, error) {
	candidates, err := findDashboardCandidates(req.DashboardIndex, req.City, 5)
	if err != nil {
		return unavailableTool("get_dashboard_context", err)
	}
	related := make([]RelatedComponent, 0)
	sources := make([]Source, 0)
	selectedIndex := ""
	if len(candidates) > 0 {
		selectedIndex = candidates[0].Index
		if components, err := getDashboardComponents(selectedIndex, req.City, 8); err == nil {
			related, sources = summarizeComponents(components)
		}
	}
	return marshalTool(toolEnvelope{
		Tool:              "get_dashboard_context",
		Sources:           sources,
		RelatedComponents: related,
		ConfidenceNotes: []string{
			"指定 dashboard_index 未命中，已改用同城市公開儀表板候選提供保守脈絡。",
		},
		Data: map[string]interface{}{
			"candidates":                candidates,
			"requested_dashboard_index": req.DashboardIndex,
			"selected_dashboard_index":  selectedIndex,
			"status":                    "degraded",
		},
	})
}

func findDashboardCandidates(query string, city string, limit int) ([]dashboardCandidate, error) {
	type row struct {
		Index string
		Name  string
		City  string
	}
	var rows []row
	db := models.DBManager.
		Table("dashboards d").
		Select("d.index, d.name, g.name as city").
		Joins("JOIN dashboard_groups dg ON d.id = dg.dashboard_id").
		Joins("JOIN groups g ON g.id = dg.group_id").
		Where("g.is_personal IS FALSE")
	if city != "" {
		db = db.Where("g.name = ?", city)
	}
	if err := db.Order("d.id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	tokens := queryTokens(query)
	candidates := make([]dashboardCandidate, 0, len(rows))
	for _, item := range rows {
		score := textScore(query, tokens, item.Index+" "+item.Name)
		if score > 0 || strings.TrimSpace(query) == "" {
			candidates = append(candidates, dashboardCandidate{
				Index: item.Index, Name: item.Name, City: item.City, Score: score,
			})
		}
	}
	if len(candidates) == 0 {
		for _, item := range rows {
			candidates = append(candidates, dashboardCandidate{
				Index: item.Index, Name: item.Name, City: item.City, Score: 0,
			})
			if len(candidates) >= limit {
				break
			}
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates, nil
}

func componentMetadataScore(query string, tokens []string, component models.CityComponent) float64 {
	text := strings.Join([]string{
		component.Index, component.Name, component.Source,
		component.ShortDesc, component.LongDesc, component.UseCase,
		component.QueryType,
	}, " ")
	return textScore(query, tokens, text)
}

func textScore(query string, tokens []string, text string) float64 {
	normalizedText := strings.ToLower(text)
	score := 0.0
	trimmed := strings.TrimSpace(strings.ToLower(query))
	if trimmed != "" && strings.Contains(normalizedText, trimmed) {
		score += 1
	}
	for _, token := range tokens {
		if token != "" && strings.Contains(normalizedText, strings.ToLower(token)) {
			score += 0.2
		}
	}
	return score
}

func queryTokens(query string) []string {
	query = strings.TrimSpace(query)
	seen := make(map[string]bool)
	tokens := make([]string, 0)
	for _, field := range strings.Fields(query) {
		addToken(&tokens, seen, field)
	}
	for _, token := range []string{
		"防災", "避難", "可達", "雨量", "水位", "示警", "交通",
		"通勤", "自行車", "長照", "高齡", "人口", "圖資", "地圖",
	} {
		if strings.Contains(query, token) {
			addToken(&tokens, seen, token)
		}
	}
	return tokens
}

func addToken(tokens *[]string, seen map[string]bool, token string) {
	token = strings.TrimSpace(token)
	if token == "" || seen[token] {
		return
	}
	seen[token] = true
	*tokens = append(*tokens, token)
}
