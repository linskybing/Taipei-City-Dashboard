package assistant

func xyRows(items []map[string]interface{}, xKey string, yKey string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		y, ok := item[yKey].(float64)
		if !ok {
			continue
		}
		out = append(out, map[string]interface{}{"x": item[xKey], "y": y})
	}
	return out
}

func pointRows(items []indexedStatPoint) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]interface{}{
			"x": item.point.X, "y": item.point.Value, "series": item.point.Series,
		})
	}
	return out
}

func anomalyRows(items []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]interface{}{
			"x": item["x"], "y": item["value"], "series": item["series"],
			"z_score": item["z_score"], "severity": item["severity"],
		})
	}
	return out
}

func forecastRows(forecast []forecastPoint, band bool) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(forecast))
	for _, point := range forecast {
		if band {
			out = append(out, map[string]interface{}{"x": point.X, "y": []float64{point.Lower, point.Upper}})
		} else {
			out = append(out, map[string]interface{}{"x": point.X, "y": point.Y})
		}
	}
	return out
}

func limitMaps(items []map[string]interface{}, limit int) []map[string]interface{} {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func asFloat(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}
