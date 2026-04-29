package assistant

import (
	"fmt"
	"sort"
	"time"
)

func buildForecastCard(ds statDataset, points []statPoint, input statToolInput) (AnalysisCard, error) {
	if len(points) < 3 {
		return AnalysisCard{}, fmt.Errorf("forecast_short_mid requires at least 3 numeric points")
	}
	horizon := intOption(input, "horizon", 3, 1, 30)
	period := intOption(input, "period", 0, 0, 366)
	forecast, method := forecastSeries(points, horizon, period)
	score := qualityScore(ds, points)
	confidence := confidenceFromQuality(score, len(points))
	findings := []string{
		fmt.Sprintf("使用 %s baseline 產生 %d 期預測。", method, horizon),
		"預測區間以歷史殘差近似，僅在近期機制不變時可作參考。",
	}
	assumptions := append(baseAssumptions(), "forecast 不代表實測值；若政策、天候或資料定義改變需重跑。")
	return statCard("forecast_short_mid", "短中期 baseline 預測已產生", findings, assumptions, ds, confidence, map[string]interface{}{
		"method":   method,
		"horizon":  horizon,
		"forecast": forecast,
	}, "approximate prediction band"), nil
}

func forecastSeries(points []statPoint, horizon int, period int) ([]forecastPoint, string) {
	sorted := sortStatPoints(points)
	if period >= 2 && len(sorted) >= period*2 {
		return seasonalNaiveForecast(sorted, horizon, period), "seasonal_naive"
	}
	return linearForecast(sorted, horizon), "linear"
}

func linearForecast(points []statPoint, horizon int) []forecastPoint {
	trend := calculateTrend(points)
	residSD := residualStdDev(points, trend)
	lastIndex := float64(len(points) - 1)
	out := make([]forecastPoint, 0, horizon)
	for step := 1; step <= horizon; step++ {
		y := trend.Intercept + trend.Slope*(lastIndex+float64(step))
		out = append(out, forecastPoint{
			X: forecastLabel(points, step),
			Y: round4(y), Lower: round4(y - 1.96*residSD), Upper: round4(y + 1.96*residSD),
		})
	}
	return out
}

func seasonalNaiveForecast(points []statPoint, horizon int, period int) []forecastPoint {
	residSD := seasonalResidualStdDev(points, period)
	out := make([]forecastPoint, 0, horizon)
	for step := 1; step <= horizon; step++ {
		base := points[len(points)-period+((step-1)%period)].Value
		out = append(out, forecastPoint{
			X: forecastLabel(points, step),
			Y: round4(base), Lower: round4(base - 1.96*residSD), Upper: round4(base + 1.96*residSD),
		})
	}
	return out
}

func residualStdDev(points []statPoint, trend trendResult) float64 {
	residuals := make([]float64, 0, len(points))
	for i, point := range points {
		expected := trend.Intercept + trend.Slope*float64(i)
		residuals = append(residuals, point.Value-expected)
	}
	return stddev(residuals)
}

func seasonalResidualStdDev(points []statPoint, period int) float64 {
	residuals := make([]float64, 0, len(points)-period)
	for i := period; i < len(points); i++ {
		residuals = append(residuals, points[i].Value-points[i-period].Value)
	}
	if len(residuals) < 2 {
		return 0
	}
	return stddev(residuals)
}

func forecastLabel(points []statPoint, step int) string {
	last := points[len(points)-1]
	if !last.HasTime {
		return fmt.Sprintf("t+%d", step)
	}
	delta := medianTimeDelta(points)
	if delta <= 0 {
		delta = 24 * time.Hour
	}
	return last.Time.Add(time.Duration(step) * delta).Format(statTimeLayout)
}

func medianTimeDelta(points []statPoint) time.Duration {
	sorted := sortStatPoints(points)
	deltas := make([]float64, 0, len(sorted)-1)
	for i := 1; i < len(sorted); i++ {
		if sorted[i].HasTime && sorted[i-1].HasTime {
			deltas = append(deltas, float64(sorted[i].Time.Sub(sorted[i-1].Time)))
		}
	}
	if len(deltas) == 0 {
		return 0
	}
	sort.Float64s(deltas)
	return time.Duration(quantileSorted(deltas, 0.5))
}
