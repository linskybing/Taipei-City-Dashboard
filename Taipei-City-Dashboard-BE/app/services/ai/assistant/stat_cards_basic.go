package assistant

import "fmt"

func buildCleanCard(ds statDataset, points []statPoint, input statToolInput) (AnalysisCard, error) {
	duplicates := countDuplicates(points)
	gaps := countTimeGaps(points)
	score := qualityScore(ds, points)
	confidence := confidenceFromQuality(score, len(points))
	findings := []string{
		fmt.Sprintf("可分析數值點 %d 筆，略過無效值 %d 筆。", len(points), ds.InvalidPoints),
		fmt.Sprintf("重複 series/x 組合 %d 筆，時間間隔疑似缺口 %d 段。", duplicates, gaps),
	}
	assumptions := []string{
		"此工具只做讀取與品質診斷，不會寫回資料或執行持久化插補。",
		"缺值檢查基於後端既有 chart parser 可見的資料點。",
	}
	recommendations := cleanRecommendations(ds.InvalidPoints, duplicates, gaps)
	return statCard("clean_impute", "資料品質已完成讀取式檢查", findings, assumptions, ds, confidence, map[string]interface{}{
		"query_type":      ds.QueryType,
		"series_name":     input.SeriesName,
		"duplicates":      duplicates,
		"time_gaps":       gaps,
		"invalid_points":  ds.InvalidPoints,
		"recommendations": recommendations,
		"visualization":   qualityVisualization(ds.InvalidPoints, duplicates, gaps, recommendations),
	}, ""), nil
}

func buildDescriptiveCard(ds statDataset, points []statPoint, input statToolInput) (AnalysisCard, error) {
	summary := describePoints(points)
	limit := intOption(input, "max_results", 5, 1, 20)
	score := qualityScore(ds, points)
	confidence := confidenceFromQuality(score, summary.Count)
	findings := []string{
		fmt.Sprintf("平均 %.4g，中位數 %.4g，標準差 %.4g。", summary.Mean, summary.Median, summary.StdDev),
		fmt.Sprintf("範圍 %.4g 至 %.4g，IQR %.4g 至 %.4g。", summary.Min, summary.Max, summary.Q1, summary.Q3),
	}
	return statCard("descriptive_report", "描述性統計已產生", findings, baseAssumptions(), ds, confidence, withVisualization(map[string]interface{}{
		"summary":         summary,
		"top_segments":    extremePoints(points, limit, true),
		"bottom_segments": extremePoints(points, limit, false),
	}, boxPlotVisualization(summary, input.SeriesName)), ""), nil
}

func buildTrendCard(ds statDataset, points []statPoint, input statToolInput) (AnalysisCard, error) {
	if len(points) < 3 {
		return AnalysisCard{}, fmt.Errorf("trend_detect requires at least 3 numeric points")
	}
	trend := calculateTrend(points)
	score := qualityScore(ds, points)
	confidence := confidenceFromQuality(score, len(points))
	if trend.Strength == "low" && confidence == "high" {
		confidence = "medium"
	}
	findings := []string{
		fmt.Sprintf("方向為 %s，趨勢強度 %s。", trend.Direction, trend.Strength),
		fmt.Sprintf("斜率 %.4g，期間首尾變化 %.4g%%。", trend.Slope, trend.PercentChange),
	}
	assumptions := append(baseAssumptions(), "趨勢偵測為描述性結果，不代表政策造成變化。")
	return statCard("trend_detect", "趨勢方向已完成估計", findings, assumptions, ds, confidence, withVisualization(map[string]interface{}{
		"trend": trend,
	}, trendVisualization(points, trend)), ""), nil
}

func statCard(
	tool string,
	headline string,
	findings []string,
	assumptions []string,
	ds statDataset,
	confidence string,
	data interface{},
	interval string,
) AnalysisCard {
	score := qualityScore(ds, ds.Points)
	return AnalysisCard{
		Tool:             tool,
		Headline:         headline,
		KeyFindings:      findings,
		Assumptions:      assumptions,
		ConfidenceLabel:  confidence,
		SourceComponents: ds.RelatedComponents,
		Uncertainty: AnalysisUncertainty{
			Interval:         emptyAs(interval, "not_applicable"),
			DataQualityScore: score,
			ConfidenceLabel:  confidence,
			CalibrationNote:  "統計結果基於既有 dashboard component data，需與資料更新頻率一併解讀。",
		},
		Data: data,
	}
}

func cleanRecommendations(invalid, duplicates, gaps int) []string {
	recs := []string{"保留原始資料，先以品質註記呈現分析限制。"}
	if invalid > 0 {
		recs = append(recs, "檢查非數值、NaN 或無限值來源，再決定是否剔除。")
	}
	if duplicates > 0 {
		recs = append(recs, "依 component 的 series/x key 檢查重複資料產生原因。")
	}
	if gaps > 0 {
		recs = append(recs, "時間序列缺口建議以不持久化方式標示或暫用線性插補做視覺參考。")
	}
	return recs
}

func baseAssumptions() []string {
	return []string{
		"輸入資料已由既有 component query_type parser 轉為官方支援格式。",
		"此結果屬描述或推論輔助，不直接產生因果宣稱。",
	}
}
