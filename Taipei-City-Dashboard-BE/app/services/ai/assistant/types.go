package assistant

import (
	"encoding/json"
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
	Sources            []Source                 `json:"sources"`
	RelatedComponents  []RelatedComponent       `json:"related_components"`
	RecommendedActions []string                 `json:"recommended_actions"`
	ConfidenceNotes    []string                 `json:"confidence_notes"`
	AnalysisCards      []AnalysisCard           `json:"analysis_cards"`
	VisualizationRefs  []VisualizationReference `json:"visualization_refs"`
}

type AnalysisUncertainty struct {
	Interval         string  `json:"interval"`
	DataQualityScore float64 `json:"data_quality_score"`
	ConfidenceLabel  string  `json:"confidence_label"`
	CalibrationNote  string  `json:"calibration_note"`
}

type AnalysisCard struct {
	Tool             string              `json:"tool"`
	Headline         string              `json:"headline"`
	KeyFindings      []string            `json:"key_findings"`
	Assumptions      []string            `json:"assumptions"`
	Uncertainty      AnalysisUncertainty `json:"uncertainty"`
	ConfidenceLabel  string              `json:"confidence_label"`
	SourceComponents []RelatedComponent  `json:"source_components"`
	Data             interface{}         `json:"data,omitempty"`
}

type VisualizationSummary struct {
	Headline        string `json:"headline"`
	Takeaway        string `json:"takeaway,omitempty"`
	AltText         string `json:"alt_text,omitempty"`
	ConfidenceLabel string `json:"confidence"`
}

type VisualizationReference struct {
	TraceID        string               `json:"trace_id"`
	ComponentID    int64                `json:"component_id,omitempty"`
	ComponentIndex string               `json:"component_index"`
	City           string               `json:"city"`
	QueryType      string               `json:"query_type,omitempty"`
	ChartConfig    json.RawMessage      `json:"chart_config,omitempty"`
	MapConfig      json.RawMessage      `json:"map_config,omitempty"`
	Summary        VisualizationSummary `json:"summary"`
	Assumptions    []string             `json:"assumptions,omitempty"`
	Uncertainty    AnalysisUncertainty  `json:"uncertainty"`
	DataSource     *Source              `json:"data_source,omitempty"`
	AuditRef       string               `json:"audit_ref"`
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
