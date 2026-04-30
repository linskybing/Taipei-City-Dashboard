package assistant

import (
	"TaipeiCityDashboardBE/app/models"
	aitools "TaipeiCityDashboardBE/app/services/ai/tools"
	"context"
	"fmt"
	"strings"
)

type toolEnvelope struct {
	Tool               string             `json:"tool"`
	Sources            []Source           `json:"sources,omitempty"`
	RelatedComponents  []RelatedComponent `json:"related_components,omitempty"`
	RecommendedActions []string           `json:"recommended_actions,omitempty"`
	ConfidenceNotes    []string           `json:"confidence_notes,omitempty"`
	Guardrails         []string           `json:"guardrails,omitempty"`
	AnalysisCards      []AnalysisCard     `json:"analysis_cards,omitempty"`
	Data               interface{}        `json:"data,omitempty"`
}

func init() {
	aitools.Register("search_components", SearchComponents)
	aitools.Register("get_component_snapshot", GetComponentSnapshot)
	aitools.Register("compare_city_components", CompareCityComponents)
	aitools.Register("get_dashboard_context", GetDashboardContext)
	aitools.Register("recommend_actions", RecommendActions)
}

func SearchComponents(ctx context.Context, args string) (string, error) {
	var input struct {
		Query          string  `json:"query"`
		Theme          string  `json:"theme"`
		City           string  `json:"city"`
		Limit          int     `json:"limit"`
		ScoreThreshold float64 `json:"score_threshold"`
	}
	if err := parseArgs(args, &input); err != nil {
		return "", err
	}
	req, err := NewContext(input.Theme, input.City, DefaultAudience, "")
	if err != nil {
		return "", err
	}
	query := strings.TrimSpace(input.Query)
	if query == "" {
		query = ThemeLabel(req.Theme)
	}
	limit := boundedInt(input.Limit, 5, 1, 10)
	score := boundedFloat(input.ScoreThreshold, 0.78, 0, 1)

	results, err := models.GetComponentByQueryVector(query, limit*2, score)
	if err != nil {
		return buildSearchComponentsFallback(query, req.City, limit)
	}
	related := filterComponents(results, req.City, limit)
	return marshalTool(toolEnvelope{
		Tool:              "search_components",
		RelatedComponents: related,
		ConfidenceNotes: []string{
			"使用 Qdrant 語意檢索公開儀表板組件 metadata。",
			"若相似度偏低，建議改以更具體的行政議題或地點描述查詢。",
		},
		Guardrails: componentSearchGuardrails(),
		Data:       map[string]interface{}{"query": query, "result_count": len(related)},
	})
}

func GetComponentSnapshot(ctx context.Context, args string) (string, error) {
	var input struct {
		ComponentID int    `json:"component_id"`
		City        string `json:"city"`
	}
	if err := parseArgs(args, &input); err != nil {
		return "", err
	}
	req, err := NewContext(DefaultTheme, input.City, DefaultAudience, "")
	if err != nil {
		return "", err
	}
	if input.ComponentID <= 0 {
		return "", fmt.Errorf("component_id is required")
	}
	component, err := getComponent(input.ComponentID, req.City)
	if err != nil {
		return unavailableTool("get_component_snapshot", err)
	}
	related := []RelatedComponent{relatedFromComponent(component, 0)}
	return marshalTool(toolEnvelope{
		Tool:              "get_component_snapshot",
		Sources:           sourcesFromComponent(component),
		RelatedComponents: related,
		ConfidenceNotes:   []string{"chart_sample 為後端以現有 query_type 解析器產生的小樣本。"},
		Data: map[string]interface{}{
			"component":    component,
			"chart_sample": getChartSample(input.ComponentID, req.City),
		},
	})
}

func CompareCityComponents(ctx context.Context, args string) (string, error) {
	var input struct {
		ComponentID    int    `json:"component_id"`
		ComponentIndex string `json:"component_index"`
	}
	if err := parseArgs(args, &input); err != nil {
		return "", err
	}
	index := strings.TrimSpace(input.ComponentIndex)
	if index == "" && input.ComponentID > 0 {
		var err error
		index, err = getComponentIndex(input.ComponentID)
		if err != nil {
			return unavailableTool("compare_city_components", err)
		}
	}
	if index == "" {
		return "", fmt.Errorf("component_id or component_index is required")
	}
	componentID, err := getComponentID(index)
	if err != nil {
		return unavailableTool("compare_city_components", err)
	}
	comparison, related, sources := compareByCity(componentID)
	return marshalTool(toolEnvelope{
		Tool:              "compare_city_components",
		Sources:           sources,
		RelatedComponents: related,
		ConfidenceNotes:   []string{"比較結果只使用同一 component index 在臺北與雙北的已設定 metadata。"},
		Data:              comparison,
	})
}

func GetDashboardContext(ctx context.Context, args string) (string, error) {
	var input struct {
		DashboardIndex string `json:"dashboard_index"`
		City           string `json:"city"`
	}
	if err := parseArgs(args, &input); err != nil {
		return "", err
	}
	req, err := NewContext(DefaultTheme, input.City, DefaultAudience, input.DashboardIndex)
	if err != nil {
		return "", err
	}
	if req.DashboardIndex == "" {
		return "", fmt.Errorf("dashboard_index is required")
	}
	components, err := getDashboardComponents(req.DashboardIndex, req.City, 8)
	if err != nil {
		return buildDashboardContextFallback(req)
	}
	related, sources := summarizeComponents(components)
	return marshalTool(toolEnvelope{
		Tool:              "get_dashboard_context",
		Sources:           sources,
		RelatedComponents: related,
		ConfidenceNotes:   []string{"只回傳前 8 個儀表板組件作為 LLM 脈絡，避免超出 token 預算。"},
		Data: map[string]interface{}{
			"dashboard_index": req.DashboardIndex,
			"components":      components,
		},
	})
}

func RecommendActions(ctx context.Context, args string) (string, error) {
	var input struct {
		Theme    string   `json:"theme"`
		City     string   `json:"city"`
		Audience string   `json:"audience"`
		Signals  []string `json:"signals"`
	}
	if err := parseArgs(args, &input); err != nil {
		return "", err
	}
	req, err := NewContext(input.Theme, input.City, input.Audience, "")
	if err != nil {
		return "", err
	}
	actions := actionsFor(req.Theme, req.Audience)
	playbook := playbookFor(req.Theme)
	return marshalTool(toolEnvelope{
		Tool:               "recommend_actions",
		RecommendedActions: actions,
		ConfidenceNotes:    []string{"建議為主題模板與已檢索訊號的決策整理，仍需由業務單位確認。"},
		Data: map[string]interface{}{
			"theme": req.Theme, "city": req.City, "audience": req.Audience,
			"signals": input.Signals, "decision_playbook": playbook,
		},
	})
}
