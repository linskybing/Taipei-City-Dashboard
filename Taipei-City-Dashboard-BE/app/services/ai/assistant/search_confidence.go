package assistant

import "strings"

type searchConfidenceBand string

const (
	searchConfidenceHigh   searchConfidenceBand = "high"
	searchConfidenceMedium searchConfidenceBand = "medium"
	searchConfidenceLow    searchConfidenceBand = "low"
)

const (
	highScoreWithAnyToken  = 0.85
	highScoreWithTwoTokens = 0.60
	mediumScoreThreshold   = 0.50
	mediumGapThreshold     = 0.1
)

type searchConfidenceAssessment struct {
	band               searchConfidenceBand
	recommended        []RelatedComponent
	analysisCandidates []RelatedComponent
	notes              []string
}

type searchEvidence struct {
	exactPhrase   bool
	tokenHits     int
	specificQuery bool
}

func assessSearchConfidence(query string, ranked []RelatedComponent, degraded bool) searchConfidenceAssessment {
	if len(ranked) == 0 {
		return searchConfidenceAssessment{band: searchConfidenceLow}
	}

	top := ranked[0]
	evidence := collectSearchEvidence(query, top)
	gap := top.Score
	if len(ranked) > 1 {
		gap = top.Score - ranked[1].Score
	}

	if degraded {
		if evidence.exactPhrase || evidence.tokenHits >= 2 {
			return searchConfidenceAssessment{
				band:        searchConfidenceMedium,
				recommended: ranked,
				notes: []string{
					"語意索引目前退化為 metadata 關鍵字比對，結果可作為線索，但建立儀表板前仍需人工確認。",
				},
			}
		}
		return searchConfidenceAssessment{
			band: searchConfidenceLow,
			notes: []string{
				"目前僅有弱關鍵字候選，暫不直接推薦組件；請補充更具體的指標、城市或資料主題。",
			},
		}
	}

	if evidence.exactPhrase || (top.Score >= highScoreWithAnyToken && evidence.tokenHits >= 1) || (top.Score >= highScoreWithTwoTokens && evidence.tokenHits >= 2) {
		return searchConfidenceAssessment{
			band:               searchConfidenceHigh,
			recommended:        ranked,
			analysisCandidates: ranked[:1],
			notes: []string{
				"檢索結果已由向量相似度與 metadata 詞項交叉比對，可作為後續工具查詢依據。",
			},
		}
	}

	if top.Score >= mediumScoreThreshold && (evidence.specificQuery || gap >= mediumGapThreshold) {
		return searchConfidenceAssessment{
			band:        searchConfidenceMedium,
			recommended: ranked,
			notes: []string{
				"檢索結果具中度信心，建議先核對組件名稱、城市與資料來源是否吻合，再建立儀表板。",
			},
		}
	}

	return searchConfidenceAssessment{
		band: searchConfidenceLow,
		notes: []string{
			"目前只找到語意相近但證據不足的候選，暫不直接推薦組件；請補充更具體的指標、城市或資料主題。",
		},
	}
}

func collectSearchEvidence(query string, component RelatedComponent) searchEvidence {
	trimmed := strings.TrimSpace(strings.ToLower(query))
	text := strings.ToLower(strings.Join([]string{component.Index, component.Name, component.QueryType}, " "))
	tokens := queryTokens(query)
	hits := 0
	for _, token := range tokens {
		if token != "" && strings.Contains(text, strings.ToLower(token)) {
			hits++
		}
	}
	return searchEvidence{
		exactPhrase:   trimmed != "" && len([]rune(trimmed)) >= 4 && strings.Contains(text, trimmed),
		tokenHits:     hits,
		specificQuery: len([]rune(trimmed)) >= 4 || len(tokens) >= 2,
	}
}

func isDegradedSearch(data map[string]interface{}) bool {
	if data == nil {
		return false
	}
	status, _ := data["status"].(string)
	if status == "degraded" {
		return true
	}
	_, ok := data["fallback"].(string)
	return ok
}
