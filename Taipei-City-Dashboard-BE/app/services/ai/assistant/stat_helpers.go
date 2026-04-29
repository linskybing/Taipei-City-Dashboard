package assistant

import (
	"fmt"
	"math"
	"sort"
)

func intOption(input statToolInput, key string, fallback, min, max int) int {
	raw, ok := input.Options[key]
	if !ok {
		return fallback
	}
	value := fallback
	switch typed := raw.(type) {
	case float64:
		value = int(typed)
	case int:
		value = typed
	}
	return boundedInt(value, fallback, min, max)
}

func stringOption(input statToolInput, key string, fallback string) string {
	raw, ok := input.Options[key].(string)
	if !ok || raw == "" {
		return fallback
	}
	return raw
}

func qualityScore(ds statDataset, points []statPoint) float64 {
	total := len(points) + ds.InvalidPoints
	if total == 0 {
		return 0
	}
	score := 1 - float64(ds.InvalidPoints)/float64(total)
	score -= math.Min(0.25, float64(countDuplicates(points))*0.01)
	score -= math.Min(0.25, float64(countTimeGaps(points))*0.02)
	if score < 0 {
		return 0
	}
	return round4(score)
}

func countDuplicates(points []statPoint) int {
	seen := make(map[string]bool)
	duplicates := 0
	for _, point := range points {
		key := point.Series + "\x00" + point.X
		if seen[key] {
			duplicates++
		}
		seen[key] = true
	}
	return duplicates
}

func countTimeGaps(points []statPoint) int {
	bySeries := groupPointsBySeries(points)
	gaps := 0
	for _, seriesPoints := range bySeries {
		sorted := sortStatPoints(seriesPoints)
		deltas := make([]float64, 0, len(sorted)-1)
		for i := 1; i < len(sorted); i++ {
			if sorted[i].HasTime && sorted[i-1].HasTime {
				deltas = append(deltas, sorted[i].Time.Sub(sorted[i-1].Time).Seconds())
			}
		}
		if len(deltas) == 0 {
			continue
		}
		sort.Float64s(deltas)
		expected := quantileSorted(deltas, 0.5)
		for _, delta := range deltas {
			if expected > 0 && delta > expected*1.5 {
				gaps++
			}
		}
	}
	return gaps
}

func groupPointsBySeries(points []statPoint) map[string][]statPoint {
	groups := make(map[string][]statPoint)
	for _, point := range points {
		groups[point.Series] = append(groups[point.Series], point)
	}
	return groups
}

func confidenceFromQuality(score float64, count int) string {
	if count < 4 || score < 0.65 {
		return "low"
	}
	if count < 12 || score < 0.85 {
		return "medium"
	}
	return "high"
}

func formatMetric(name string, value float64) string {
	return fmt.Sprintf("%s = %.4g", name, value)
}

func extremePoints(points []statPoint, limit int, descending bool) []map[string]interface{} {
	sorted := append([]statPoint(nil), points...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if descending {
			return sorted[i].Value > sorted[j].Value
		}
		return sorted[i].Value < sorted[j].Value
	})
	if len(sorted) > limit {
		sorted = sorted[:limit]
	}
	out := make([]map[string]interface{}, 0, len(sorted))
	for _, point := range sorted {
		out = append(out, map[string]interface{}{
			"series": point.Series,
			"x":      point.X,
			"value":  point.Value,
		})
	}
	return out
}
