package assistant

import "fmt"

const (
	maxVisualizationPoints    = 30
	maxVisualizationItems     = 12
	maxVisualizationAnomalies = 10
)

func withVisualization(data map[string]interface{}, viz map[string]interface{}) map[string]interface{} {
	if viz != nil {
		data["visualization"] = viz
	}
	return data
}

func qualityVisualization(invalid, duplicates, gaps int, recommendations []string) map[string]interface{} {
	return map[string]interface{}{
		"kind": "quality_summary",
		"series": []map[string]interface{}{
			{"label": "無效值", "value": invalid},
			{"label": "重複", "value": duplicates},
			{"label": "時間缺口", "value": gaps},
		},
		"annotations": recommendations,
		"caption":     "資料品質診斷適合以指標摘要呈現，不適合用長條圖暗示三者可直接比較。",
	}
}

func boxPlotVisualization(summary statSummary, seriesName string) map[string]interface{} {
	return map[string]interface{}{
		"kind": "box_plot",
		"series": []map[string]interface{}{{
			"x": emptyAs(seriesName, "資料分布"),
			"y": []float64{summary.Min, summary.Q1, summary.Median, summary.Q3, summary.Max},
		}},
		"annotations": map[string]interface{}{
			"mean": summary.Mean, "stddev": summary.StdDev, "count": summary.Count,
		},
		"caption": "盒鬚圖呈現 min、Q1、中位數、Q3、max；平均與標準差作為輔助註記。",
	}
}

func trendVisualization(points []statPoint, trend trendResult) map[string]interface{} {
	observed, fitted := trendSeries(points, trend)
	return map[string]interface{}{
		"kind": "trend_line",
		"series": []map[string]interface{}{
			{"name": "觀測值", "data": observed},
			{"name": "趨勢線", "data": fitted},
		},
		"annotations": map[string]interface{}{
			"direction": trend.Direction, "strength": trend.Strength, "r2": trend.R2,
			"percent_change": trend.PercentChange, "slope": trend.Slope,
		},
		"caption": fmt.Sprintf("趨勢以觀測點與 fitted line 呈現；首尾變化 %.4g%%，R2 %.4g。", trend.PercentChange, trend.R2),
	}
}

func seasonalVisualization(buckets []map[string]interface{}, residuals []map[string]interface{}, period int) map[string]interface{} {
	limitedBuckets := limitMaps(buckets, maxVisualizationItems)
	return map[string]interface{}{
		"kind": "seasonal_profile",
		"series": []map[string]interface{}{{
			"name": "bucket 平均", "data": xyRows(limitedBuckets, "bucket", "mean"),
		}},
		"annotations": map[string]interface{}{
			"period": period, "largest_abs_residuals": limitMaps(residuals, maxVisualizationAnomalies),
		},
		"caption": "季節 profile 保留 bucket 原始順序，避免排序後誤讀為排名圖。",
	}
}

func anomalyVisualization(points []statPoint, anomalies []map[string]interface{}) map[string]interface{} {
	values := statValues(points)
	summary := describePoints(points)
	iqr := summary.Q3 - summary.Q1
	return map[string]interface{}{
		"kind": "anomaly_context",
		"series": []map[string]interface{}{
			{"name": "觀測值", "data": pointRows(boundedIndexedPoints(points, maxVisualizationPoints))},
			{"name": "候選異常", "data": anomalyRows(limitMaps(anomalies, maxVisualizationAnomalies))},
		},
		"annotations": map[string]interface{}{
			"mean": mean(values), "lower_fence": summary.Q1 - 1.5*iqr, "upper_fence": summary.Q3 + 1.5*iqr,
		},
		"caption": "異常點必須放回觀測序列與 IQR fence 脈絡中解讀，不能只看排名長條。",
	}
}

func effectIntervalVisualization(result map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"kind": "effect_interval",
		"series": []map[string]interface{}{{
			"x":        "平均差 A-B",
			"y":        []float64{asFloat(result["ci_low"]), asFloat(result["ci_high"])},
			"estimate": asFloat(result["difference"]),
		}},
		"annotations": map[string]interface{}{"zero_line": 0, "p_value": result["p_value_normal_approx"]},
		"caption":     "兩組比較以效果量信賴區間與零線呈現，避免只比較兩根平均值。",
	}
}

func contingencyVisualization(table [2][2]int, names []string, p float64) map[string]interface{} {
	series := make([]map[string]interface{}, 0, 2)
	for i, row := range table {
		name := fmt.Sprintf("系列 %d", i+1)
		if i < len(names) {
			name = names[i]
		}
		series = append(series, map[string]interface{}{"name": name, "data": []map[string]interface{}{
			{"x": "欄 1", "y": row[0]}, {"x": "欄 2", "y": row[1]},
		}})
	}
	return map[string]interface{}{
		"kind": "contingency_heatmap", "series": series,
		"annotations": map[string]interface{}{"fisher_p_value": p},
		"caption":     "2x2 類別檢定適合用 heatmap 顯示列欄交叉計數，p 值只作輔助解讀。",
	}
}

func forecastVisualization(points []statPoint, forecast []forecastPoint, method string) map[string]interface{} {
	return map[string]interface{}{
		"kind": "forecast_band",
		"series": []map[string]interface{}{
			{"name": "觀測值", "data": pointRows(boundedTailPoints(points, maxVisualizationPoints))},
			{"name": "預測", "data": forecastRows(forecast, false)},
			{"name": "預測區間", "data": forecastRows(forecast, true)},
		},
		"annotations": map[string]interface{}{"method": method, "horizon": len(forecast)},
		"caption":     "預測以中心線加 prediction band 呈現；區間代表模型不確定性，不是兩條獨立序列。",
	}
}

type indexedStatPoint struct {
	point statPoint
	index int
}

func trendSeries(points []statPoint, trend trendResult) ([]map[string]interface{}, []map[string]interface{}) {
	sorted := sortStatPoints(points)
	indexed := boundedIndexedFromSorted(sorted, maxVisualizationPoints)
	observed := pointRows(indexed)
	fitted := make([]map[string]interface{}, 0, len(indexed))
	for _, item := range indexed {
		x := float64(item.index)
		if item.point.HasTime {
			x = item.point.Time.Sub(sorted[0].Time).Hours() / 24
		}
		fitted = append(fitted, map[string]interface{}{"x": item.point.X, "y": round4(trend.Intercept + trend.Slope*x)})
	}
	return observed, fitted
}

func boundedIndexedPoints(points []statPoint, limit int) []indexedStatPoint {
	return boundedIndexedFromSorted(sortStatPoints(points), limit)
}

func boundedTailPoints(points []statPoint, limit int) []indexedStatPoint {
	sorted := sortStatPoints(points)
	if len(sorted) > limit {
		offset := len(sorted) - limit
		out := make([]indexedStatPoint, 0, limit)
		for i, point := range sorted[offset:] {
			out = append(out, indexedStatPoint{point: point, index: offset + i})
		}
		return out
	}
	return boundedIndexedFromSorted(sorted, limit)
}

func boundedIndexedFromSorted(sorted []statPoint, limit int) []indexedStatPoint {
	if len(sorted) <= limit {
		out := make([]indexedStatPoint, 0, len(sorted))
		for i, point := range sorted {
			out = append(out, indexedStatPoint{point: point, index: i})
		}
		return out
	}
	out := make([]indexedStatPoint, 0, limit)
	for i := 0; i < limit; i++ {
		index := i * (len(sorted) - 1) / (limit - 1)
		out = append(out, indexedStatPoint{point: sorted[index], index: index})
	}
	return out
}
