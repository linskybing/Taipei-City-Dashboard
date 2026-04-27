package assistant

import "encoding/json"

type ToolEvent struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Loop      int    `json:"loop"`
	LatencyMS int    `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

func BuildMetadata(
	base map[string]interface{},
	toolNames []string,
	toolOutputs []string,
	toolEvents []ToolEvent,
) string {
	metadata := make(map[string]interface{})
	for key, value := range base {
		metadata[key] = value
	}
	metadata["tool_names"] = toolNames
	metadata["tool_events"] = toolEvents

	extras := mergeToolOutputs(toolOutputs)
	metadata["sources"] = extras.Sources
	metadata["related_components"] = extras.RelatedComponents
	metadata["recommended_actions"] = extras.RecommendedActions
	metadata["confidence_notes"] = extras.ConfidenceNotes

	data, err := json.Marshal(metadata)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func ResponseExtrasFromMetadata(raw string) ResponseExtras {
	var rawMap map[string]json.RawMessage
	var extras ResponseExtras
	if json.Unmarshal([]byte(raw), &rawMap) != nil {
		return extras
	}
	json.Unmarshal(rawMap["sources"], &extras.Sources)
	json.Unmarshal(rawMap["related_components"], &extras.RelatedComponents)
	json.Unmarshal(rawMap["recommended_actions"], &extras.RecommendedActions)
	json.Unmarshal(rawMap["confidence_notes"], &extras.ConfidenceNotes)
	return extras
}

func mergeToolOutputs(outputs []string) ResponseExtras {
	extras := ResponseExtras{}
	sourceSeen := make(map[string]bool)
	componentSeen := make(map[string]bool)
	actionSeen := make(map[string]bool)
	noteSeen := make(map[string]bool)

	for _, output := range outputs {
		var envelope toolEnvelope
		if json.Unmarshal([]byte(output), &envelope) != nil {
			continue
		}
		for _, source := range envelope.Sources {
			key := source.Name + source.URL
			if !sourceSeen[key] {
				extras.Sources = append(extras.Sources, source)
				sourceSeen[key] = true
			}
		}
		for _, component := range envelope.RelatedComponents {
			key := component.Index + component.City
			if !componentSeen[key] {
				extras.RelatedComponents = append(extras.RelatedComponents, component)
				componentSeen[key] = true
			}
		}
		appendUniqueStrings(&extras.RecommendedActions, envelope.RecommendedActions, actionSeen)
		appendUniqueStrings(&extras.ConfidenceNotes, envelope.ConfidenceNotes, noteSeen)
	}
	return extras
}

func appendUniqueStrings(target *[]string, values []string, seen map[string]bool) {
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		*target = append(*target, value)
		seen[value] = true
	}
}
