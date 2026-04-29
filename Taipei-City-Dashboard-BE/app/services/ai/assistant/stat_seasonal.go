package assistant

import (
	"fmt"
	"math"
	"sort"
)

func buildSeasonalCard(ds statDataset, points []statPoint, input statToolInput) (AnalysisCard, error) {
	period := intOption(input, "period", 7, 2, 366)
	if len(points) < period*2 {
		return AnalysisCard{}, fmt.Errorf("seasonal_decompose requires at least two full periods")
	}
	buckets, residuals := seasonalProfile(points, period)
	score := qualityScore(ds, points)
	confidence := confidenceFromQuality(score, len(points))
	findings := []string{
		fmt.Sprintf("以 period=%d 建立輕量季節 bucket，共 %d 個 bucket。", period, len(buckets)),
		fmt.Sprintf("最大平均 bucket 為 %s，最小平均 bucket 為 %s。", buckets[len(buckets)-1]["bucket"], buckets[0]["bucket"]),
	}
	assumptions := append(baseAssumptions(), "季節分解採 bucket 平均近似，不使用 STL 或外部套件。")
	return statCard("seasonal_decompose", "季節 bucket 與殘差偏離已估計", findings, assumptions, ds, confidence, map[string]interface{}{
		"period":                period,
		"bucket_profile":        buckets,
		"largest_abs_residuals": residuals,
	}, ""), nil
}

func buildAnomalyCard(ds statDataset, points []statPoint, input statToolInput) (AnalysisCard, error) {
	if len(points) < 4 {
		return AnalysisCard{}, fmt.Errorf("anomaly_detect requires at least 4 numeric points")
	}
	limit := intOption(input, "max_results", 10, 1, 20)
	anomalies := anomalyPoints(points, limit)
	score := qualityScore(ds, points)
	confidence := confidenceFromQuality(score, len(points))
	findings := []string{
		fmt.Sprintf("偵測到 %d 筆超出 z-score/IQR 基線的候選異常。", len(anomalies)),
		"異常只代表偏離既有分布，需由業務事件或資料品質再確認。",
	}
	assumptions := append(baseAssumptions(), "異常偵測使用整體分布基線；概念漂移時需重新檢查。")
	return statCard("anomaly_detect", "候選異常點已完成排序", findings, assumptions, ds, confidence, map[string]interface{}{
		"anomalies": anomalies,
	}, ""), nil
}

func seasonalProfile(points []statPoint, period int) ([]map[string]interface{}, []map[string]interface{}) {
	sorted := sortStatPoints(points)
	bucketValues := make([][]float64, period)
	for i, point := range sorted {
		bucket := i % period
		bucketValues[bucket] = append(bucketValues[bucket], point.Value)
	}
	means := make([]float64, period)
	buckets := make([]map[string]interface{}, 0, period)
	for i, values := range bucketValues {
		means[i] = mean(values)
		buckets = append(buckets, map[string]interface{}{
			"bucket": fmt.Sprintf("%d", i+1),
			"mean":   round4(means[i]),
			"count":  len(values),
		})
	}
	sort.SliceStable(buckets, func(i, j int) bool {
		return buckets[i]["mean"].(float64) < buckets[j]["mean"].(float64)
	})
	residuals := make([]map[string]interface{}, 0, len(sorted))
	for i, point := range sorted {
		resid := point.Value - means[i%period]
		residuals = append(residuals, map[string]interface{}{
			"x":            point.X,
			"series":       point.Series,
			"bucket":       (i % period) + 1,
			"residual":     round4(resid),
			"abs_residual": round4(math.Abs(resid)),
		})
	}
	sort.SliceStable(residuals, func(i, j int) bool {
		return residuals[i]["abs_residual"].(float64) > residuals[j]["abs_residual"].(float64)
	})
	if len(residuals) > 10 {
		residuals = residuals[:10]
	}
	return buckets, residuals
}

func anomalyPoints(points []statPoint, limit int) []map[string]interface{} {
	values := statValues(points)
	sortedValues := append([]float64(nil), values...)
	sort.Float64s(sortedValues)
	avg, sd := mean(values), stddev(values)
	q1, q3 := quantileSorted(sortedValues, 0.25), quantileSorted(sortedValues, 0.75)
	iqr := q3 - q1
	rows := make([]map[string]interface{}, 0)
	for _, point := range points {
		z := 0.0
		if sd > 0 {
			z = (point.Value - avg) / sd
		}
		outsideIQR := iqr > 0 && (point.Value < q1-1.5*iqr || point.Value > q3+1.5*iqr)
		if math.Abs(z) < 2.5 && !outsideIQR {
			continue
		}
		severity := "medium"
		if math.Abs(z) >= 3.5 {
			severity = "high"
		}
		rows = append(rows, map[string]interface{}{
			"x": point.X, "series": point.Series, "value": point.Value,
			"z_score": round4(z), "severity": severity,
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return math.Abs(rows[i]["z_score"].(float64)) > math.Abs(rows[j]["z_score"].(float64))
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}
