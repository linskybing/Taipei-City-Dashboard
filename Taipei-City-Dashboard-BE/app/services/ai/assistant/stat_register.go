package assistant

import aitools "TaipeiCityDashboardBE/app/services/ai/tools"

func init() {
	aitools.Register("clean_impute", CleanImpute)
	aitools.Register("descriptive_report", DescriptiveReport)
	aitools.Register("trend_detect", TrendDetect)
	aitools.Register("seasonal_decompose", SeasonalDecompose)
	aitools.Register("anomaly_detect", AnomalyDetect)
	aitools.Register("hypothesis_test", HypothesisTest)
	aitools.Register("forecast_short_mid", ForecastShortMid)
}
