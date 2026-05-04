package assistant

import (
	"math"
	"sort"
)

func describePoints(points []statPoint) statSummary {
	values := statValues(points)
	sort.Float64s(values)
	return statSummary{
		Count:  len(values),
		Min:    values[0],
		Max:    values[len(values)-1],
		Mean:   mean(values),
		Median: quantileSorted(values, 0.5),
		StdDev: stddev(values),
		Q1:     quantileSorted(values, 0.25),
		Q3:     quantileSorted(values, 0.75),
	}
}

func calculateTrend(points []statPoint) trendResult {
	sorted := sortStatPoints(points)
	n := float64(len(sorted))
	var sumX, sumY, sumXY, sumXX, sumYY float64
	for i, point := range sorted {
		x := float64(i)
		if point.HasTime {
			x = point.Time.Sub(sorted[0].Time).Hours() / 24
		}
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
		sumYY += y * y
	}
	denom := n*sumXX - sumX*sumX
	slope := 0.0
	if denom != 0 {
		slope = (n*sumXY - sumX*sumY) / denom
	}
	intercept := (sumY - slope*sumX) / n
	r := correlation(n, sumX, sumY, sumXY, sumXX, sumYY)
	return trendResult{
		Direction:     trendDirection(slope),
		Strength:      trendStrength(r),
		Slope:         round4(slope),
		Intercept:     round4(intercept),
		R:             round4(r),
		R2:            round4(r * r),
		PercentChange: round4(percentChange(sorted[0].Value, sorted[len(sorted)-1].Value)),
	}
}

func statValues(points []statPoint) []float64 {
	values := make([]float64, 0, len(points))
	for _, point := range points {
		values = append(values, point.Value)
	}
	return values
}

func sortStatPoints(points []statPoint) []statPoint {
	sorted := append([]statPoint(nil), points...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].HasTime && sorted[j].HasTime {
			return sorted[i].Time.Before(sorted[j].Time)
		}
		if sorted[i].Series == sorted[j].Series {
			return sorted[i].X < sorted[j].X
		}
		return sorted[i].Series < sorted[j].Series
	})
	return sorted
}

func mean(values []float64) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func variance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	avg := mean(values)
	sum := 0.0
	for _, value := range values {
		diff := value - avg
		sum += diff * diff
	}
	return sum / float64(len(values)-1)
}

func stddev(values []float64) float64 {
	return math.Sqrt(variance(values))
}

func quantileSorted(values []float64, q float64) float64 {
	if len(values) == 1 {
		return values[0]
	}
	pos := q * float64(len(values)-1)
	lower := int(math.Floor(pos))
	upper := int(math.Ceil(pos))
	if lower == upper {
		return values[lower]
	}
	weight := pos - float64(lower)
	return values[lower]*(1-weight) + values[upper]*weight
}

func correlation(n, sumX, sumY, sumXY, sumXX, sumYY float64) float64 {
	denom := math.Sqrt((n*sumXX - sumX*sumX) * (n*sumYY - sumY*sumY))
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

func percentChange(first, last float64) float64 {
	if first == 0 {
		return 0
	}
	return ((last - first) / math.Abs(first)) * 100
}

func trendDirection(slope float64) string {
	if math.Abs(slope) < 0.000001 {
		return "flat"
	}
	if slope > 0 {
		return "increasing"
	}
	return "decreasing"
}

func trendStrength(r float64) string {
	absR := math.Abs(r)
	if absR >= 0.75 {
		return "high"
	}
	if absR >= 0.45 {
		return "medium"
	}
	return "low"
}

func round4(value float64) float64 {
	return math.Round(value*10000) / 10000
}
