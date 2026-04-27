package assistant

import (
	"fmt"
	"strings"
)

const (
	DefaultTheme    = "auto"
	DefaultCity     = "metrotaipei"
	DefaultAudience = "government"
)

type RequestContext struct {
	Theme          string `json:"theme"`
	City           string `json:"city"`
	Audience       string `json:"audience"`
	DashboardIndex string `json:"dashboard_index,omitempty"`
}

type Source struct {
	Name string   `json:"name"`
	Type string   `json:"type,omitempty"`
	URL  string   `json:"url,omitempty"`
	URLs []string `json:"urls,omitempty"`
}

type RelatedComponent struct {
	ID        int64   `json:"id"`
	Index     string  `json:"index"`
	Name      string  `json:"name"`
	City      string  `json:"city"`
	QueryType string  `json:"query_type,omitempty"`
	Score     float64 `json:"score,omitempty"`
}

type ResponseExtras struct {
	Sources            []Source           `json:"sources"`
	RelatedComponents  []RelatedComponent `json:"related_components"`
	RecommendedActions []string           `json:"recommended_actions"`
	ConfidenceNotes    []string           `json:"confidence_notes"`
}

func NewContext(theme, city, audience, dashboardIndex string) (RequestContext, error) {
	normalizedTheme, err := normalize(theme, validThemes(), DefaultTheme, "theme")
	if err != nil {
		return RequestContext{}, err
	}
	normalizedCity, err := normalize(city, validCities(), DefaultCity, "city")
	if err != nil {
		return RequestContext{}, err
	}
	normalizedAudience, err := normalize(audience, validAudiences(), DefaultAudience, "audience")
	if err != nil {
		return RequestContext{}, err
	}
	return RequestContext{
		Theme:          normalizedTheme,
		City:           normalizedCity,
		Audience:       normalizedAudience,
		DashboardIndex: strings.TrimSpace(dashboardIndex),
	}, nil
}

func (c RequestContext) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"theme":             c.Theme,
		"city":              c.City,
		"audience":          c.Audience,
		"dashboard_index":   c.DashboardIndex,
		"decision_playbook": playbookFor(c.Theme),
	}
}

func normalize(value string, allowed map[string]bool, fallback string, field string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		value = fallback
	}
	if !allowed[value] {
		return "", fmt.Errorf("invalid %s: %s", field, value)
	}
	return value, nil
}

func validThemes() map[string]bool {
	return map[string]bool{
		"auto": true, "commuting": true, "disaster": true,
		"environment": true, "health": true, "labor": true, "culture": true,
	}
}

func validCities() map[string]bool {
	return map[string]bool{"taipei": true, "metrotaipei": true}
}

func validAudiences() map[string]bool {
	return map[string]bool{"government": true, "public": true}
}
