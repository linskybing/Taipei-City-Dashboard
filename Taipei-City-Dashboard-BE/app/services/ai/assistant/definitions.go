package assistant

import "github.com/tmc/langchaingo/llms"

func ToolDefinitions() []llms.Tool {
	definitions := []llms.Tool{
		defineTool("search_components", "Search dashboard components by a natural-language policy or city signal query.", map[string]interface{}{
			"query":           stringSchema("Natural-language query to search for matching dashboard components."),
			"theme":           enumSchema(validThemeValues(), "Hackathon theme."),
			"city":            enumSchema(validCityValues(), "City scope."),
			"limit":           integerSchema("Maximum result count, 1 to 10."),
			"score_threshold": numberSchema("Similarity threshold from 0 to 1."),
		}, []string{"query"}),
		defineTool("get_component_snapshot", "Get metadata and a small chart sample for one component.", map[string]interface{}{
			"component_id": integerSchema("Component numeric ID."),
			"city":         enumSchema(validCityValues(), "City scope."),
		}, []string{"component_id"}),
		defineTool("compare_city_components", "Compare Taipei and Metro Taipei metadata for a component index.", map[string]interface{}{
			"component_id":    integerSchema("Component numeric ID."),
			"component_index": stringSchema("Component English index if known."),
		}, []string{}),
		defineTool("get_dashboard_context", "Get component context for a dashboard index.", map[string]interface{}{
			"dashboard_index": stringSchema("Dashboard index."),
			"city":            enumSchema(validCityValues(), "City scope."),
		}, []string{"dashboard_index"}),
		defineTool("recommend_actions", "Generate bounded action suggestions for a theme and audience.", map[string]interface{}{
			"theme":    enumSchema(validThemeValues(), "Hackathon theme."),
			"city":     enumSchema(validCityValues(), "City scope."),
			"audience": enumSchema(validAudienceValues(), "Decision audience."),
			"signals":  arraySchema("Dashboard signals or facts already found."),
		}, []string{"theme", "city", "audience"}),
	}
	return append(definitions, statisticalToolDefinitions()...)
}

func defineTool(name string, description string, properties map[string]interface{}, required []string) llms.Tool {
	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        name,
			Description: description,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		},
	}
}

func stringSchema(description string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": description}
}

func numberSchema(description string) map[string]interface{} {
	return map[string]interface{}{"type": "number", "description": description}
}

func integerSchema(description string) map[string]interface{} {
	return map[string]interface{}{"type": "integer", "description": description}
}

func arraySchema(description string) map[string]interface{} {
	return map[string]interface{}{
		"type": "array", "description": description,
		"items": map[string]interface{}{"type": "string"},
	}
}

func enumSchema(values []string, description string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "enum": values, "description": description}
}

func validThemeValues() []string {
	return []string{"auto", "commuting", "disaster", "environment", "health", "labor", "culture"}
}

func validCityValues() []string {
	return []string{"taipei", "metrotaipei"}
}

func validAudienceValues() []string {
	return []string{"government", "public"}
}
