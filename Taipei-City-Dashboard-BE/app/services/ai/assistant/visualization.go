package assistant

import (
	"TaipeiCityDashboardBE/app/models"
	"fmt"
)

func BuildVisualizationRefs(extras ResponseExtras, auditRef string) []VisualizationReference {
	components := collectVisualizationComponents(extras)
	refs := make([]VisualizationReference, 0, len(components))
	for i, related := range components {
		component, ok := loadVisualizationComponent(related)
		card := matchingAnalysisCard(extras.AnalysisCards, related)
		if ok {
			related = relatedFromComponent(component, related.Score)
		}
		ref := visualizationReference(related, component, ok, card, auditRef, i+1)
		if ref.ComponentIndex != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}

func collectVisualizationComponents(extras ResponseExtras) []RelatedComponent {
	seen := make(map[string]bool)
	components := make([]RelatedComponent, 0, len(extras.RelatedComponents))
	for _, component := range extras.RelatedComponents {
		appendVisualizationComponent(&components, seen, component)
	}
	for _, card := range extras.AnalysisCards {
		for _, component := range card.SourceComponents {
			appendVisualizationComponent(&components, seen, component)
		}
	}
	return components
}

func appendVisualizationComponent(
	components *[]RelatedComponent,
	seen map[string]bool,
	component RelatedComponent,
) {
	key := component.Index + ":" + component.City
	if component.Index == "" || component.City == "" || seen[key] {
		return
	}
	*components = append(*components, component)
	seen[key] = true
}

func loadVisualizationComponent(related RelatedComponent) (models.CityComponent, bool) {
	if related.ID <= 0 || models.DBManager == nil {
		return models.CityComponent{}, false
	}
	component, err := getComponent(int(related.ID), related.City)
	return component, err == nil
}

func matchingAnalysisCard(cards []AnalysisCard, related RelatedComponent) *AnalysisCard {
	for i := range cards {
		for _, component := range cards[i].SourceComponents {
			if sameRelatedComponent(component, related) {
				return &cards[i]
			}
		}
	}
	if len(cards) == 1 {
		return &cards[0]
	}
	return nil
}

func sameRelatedComponent(a RelatedComponent, b RelatedComponent) bool {
	if a.ID > 0 && b.ID > 0 && a.ID == b.ID && a.City == b.City {
		return true
	}
	return a.Index != "" && a.Index == b.Index && a.City == b.City
}

func visualizationReference(
	related RelatedComponent,
	component models.CityComponent,
	hasComponent bool,
	card *AnalysisCard,
	auditRef string,
	seq int,
) VisualizationReference {
	summary := VisualizationSummary{
		Headline:        related.Name,
		ConfidenceLabel: "medium",
	}
	uncertainty := AnalysisUncertainty{
		Interval:         "not_applicable",
		DataQualityScore: 0.7,
		ConfidenceLabel:  "medium",
		CalibrationNote:  "此引用只連結既有 dashboard component data，不承載原始資料列。",
	}
	assumptions := []string{"資料由既有 component query 與 chart_config 呈現。"}
	if hasComponent {
		summary.Headline = component.Name
		summary.Takeaway = emptyAs(component.ShortDesc, component.LongDesc)
		summary.AltText = component.UseCase
	}
	if card != nil {
		summary.Headline = emptyAs(card.Headline, summary.Headline)
		summary.ConfidenceLabel = emptyAs(card.ConfidenceLabel, summary.ConfidenceLabel)
		assumptions = card.Assumptions
		uncertainty = card.Uncertainty
	}
	source := visualizationSource(component, hasComponent)
	return VisualizationReference{
		TraceID:        traceID(auditRef, seq),
		ComponentID:    related.ID,
		ComponentIndex: related.Index,
		City:           related.City,
		QueryType:      emptyAs(related.QueryType, component.QueryType),
		ChartConfig:    component.ChartConfig,
		MapConfig:      component.MapConfig,
		Summary:        summary,
		Assumptions:    assumptions,
		Uncertainty:    uncertainty,
		DataSource:     source,
		AuditRef:       auditRef,
	}
}

func visualizationSource(component models.CityComponent, hasComponent bool) *Source {
	if !hasComponent {
		return nil
	}
	sources := sourcesFromComponent(component)
	if len(sources) == 0 {
		return nil
	}
	return &sources[0]
}

func traceID(auditRef string, seq int) string {
	if auditRef == "" {
		return ""
	}
	return fmt.Sprintf("%s:visualization:%d", auditRef, seq)
}
