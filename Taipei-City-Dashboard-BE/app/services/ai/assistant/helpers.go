package assistant

import (
	"TaipeiCityDashboardBE/app/models"
	"encoding/json"
)

func parseArgs(args string, value interface{}) error {
	return json.Unmarshal([]byte(args), value)
}

func marshalTool(value toolEnvelope) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func boundedInt(value int, fallback int, min int, max int) int {
	if value == 0 {
		value = fallback
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func boundedFloat(value float64, fallback float64, min float64, max float64) float64 {
	if value == 0 {
		value = fallback
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

type filteredComponents struct {
	related                 []RelatedComponent
	cityResolutionDropCount int
	cityResolutionDropPreview []cityResolutionDropPreview
}

type cityResolutionDropPreview struct {
	Index string `json:"index"`
	Name  string `json:"name,omitempty"`
	City  string `json:"city,omitempty"`
}

func filterComponents(items []models.CityComponentScore, city string, limit int) filteredComponents {
	return filterComponentsWithResolver(items, city, limit, getComponent)
}

func filterComponentsWithResolver(
	items []models.CityComponentScore,
	city string,
	limit int,
	resolve func(int, string) (models.CityComponent, error),
) filteredComponents {
	seen := make(map[string]RelatedComponent)
	order := make([]string, 0, limit)
	previewKeys := make(map[string]bool)
	dropCount := 0
	dropPreview := make([]cityResolutionDropPreview, 0, 2)
	for _, item := range items {
		related := RelatedComponent{
			ID: item.ID, Index: item.Index, Name: item.Name, City: item.City, Score: item.Score,
		}
		if city != "" && item.City != city {
			component, err := resolve(int(item.ID), city)
			if err != nil {
				dropCount++
				previewKey := item.Index + ":" + item.City
				if !previewKeys[previewKey] && len(dropPreview) < 2 {
					previewKeys[previewKey] = true
					dropPreview = append(dropPreview, cityResolutionDropPreview{
						Index: item.Index,
						Name:  item.Name,
						City:  item.City,
					})
				}
				continue
			}
			related = relatedFromComponent(component, item.Score)
		}
		key := related.Index + ":" + related.City
		if _, ok := seen[key]; !ok {
			order = append(order, key)
		}
		seen[key] = related
		if len(order) >= limit {
			break
		}
	}
	related := make([]RelatedComponent, 0, len(order))
	for _, key := range order {
		related = append(related, seen[key])
	}
	return filteredComponents{
		related:                   related,
		cityResolutionDropCount:   dropCount,
		cityResolutionDropPreview: dropPreview,
	}
}

func relatedFromComponent(component models.CityComponent, score float64) RelatedComponent {
	return RelatedComponent{
		ID:        component.ID,
		Index:     component.Index,
		Name:      component.Name,
		City:      component.City,
		QueryType: component.QueryType,
		Score:     score,
	}
}

func sourcesFromComponent(component models.CityComponent) []Source {
	source := Source{Name: component.Source, Type: "open_data"}
	for _, link := range component.Links {
		if link != "" {
			source.URLs = append(source.URLs, link)
		}
	}
	if len(source.URLs) == 1 {
		source.URL = source.URLs[0]
	}
	if source.Name == "" && len(source.URLs) == 0 {
		return nil
	}
	return []Source{source}
}

func summarizeComponents(components []models.CityComponent) ([]RelatedComponent, []Source) {
	related := make([]RelatedComponent, 0, len(components))
	sourceMap := make(map[string]Source)
	for _, component := range components {
		related = append(related, relatedFromComponent(component, 0))
		for _, source := range sourcesFromComponent(component) {
			key := source.Name + source.URL
			sourceMap[key] = source
		}
	}
	sources := make([]Source, 0, len(sourceMap))
	for _, source := range sourceMap {
		sources = append(sources, source)
	}
	return related, sources
}

func compareByCity(componentID int) (map[string]interface{}, []RelatedComponent, []Source) {
	result := make(map[string]interface{})
	related := make([]RelatedComponent, 0, 2)
	sourceMap := make(map[string]Source)
	for _, city := range []string{"taipei", "metrotaipei"} {
		component, err := getComponent(componentID, city)
		if err != nil {
			result[city] = map[string]string{"error": err.Error()}
			continue
		}
		result[city] = component
		related = append(related, relatedFromComponent(component, 0))
		for _, source := range sourcesFromComponent(component) {
			sourceMap[source.Name+source.URL] = source
		}
	}
	sources := make([]Source, 0, len(sourceMap))
	for _, source := range sourceMap {
		sources = append(sources, source)
	}
	return result, related, sources
}
