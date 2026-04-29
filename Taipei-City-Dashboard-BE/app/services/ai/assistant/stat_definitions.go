package assistant

import "github.com/tmc/langchaingo/llms"

func statisticalToolDefinitions() []llms.Tool {
	props := statToolProperties()
	return []llms.Tool{
		defineTool("clean_impute", "Profile component data quality and recommend non-persistent cleaning or imputation actions.", props, []string{"component_id"}),
		defineTool("descriptive_report", "Summarize numeric component data with distribution and top/bottom segment statistics.", props, []string{"component_id"}),
		defineTool("trend_detect", "Detect component trend direction, slope, percent change, and trend strength.", props, []string{"component_id"}),
		defineTool("seasonal_decompose", "Estimate lightweight seasonal buckets and residual deviations without external packages.", props, []string{"component_id"}),
		defineTool("anomaly_detect", "Detect point anomalies using z-score and IQR baselines with bounded results.", props, []string{"component_id"}),
		defineTool("hypothesis_test", "Run guarded Welch or 2x2 categorical tests when component data shape is suitable.", props, []string{"component_id"}),
		defineTool("forecast_short_mid", "Create short or medium horizon forecasts using naive, seasonal-naive, or linear baselines.", props, []string{"component_id"}),
	}
}

func statToolProperties() map[string]interface{} {
	return map[string]interface{}{
		"component_id": integerSchema("Dashboard component numeric ID to analyze."),
		"city":         enumSchema(validCityValues(), "City scope. Defaults to the current assistant context when omitted."),
		"time_range": objectSchema("Optional ISO-8601 time range for time-filtered component queries.", map[string]interface{}{
			"start": stringSchema("Start timestamp or date."),
			"end":   stringSchema("End timestamp or date."),
		}),
		"series_name": stringSchema("Optional series name to focus analysis on one component series."),
		"options": objectSchema("Tool-specific bounded options such as period, horizon, max_results, or test.", map[string]interface{}{
			"period":      integerSchema("Seasonal period, 2 to 366."),
			"horizon":     integerSchema("Forecast horizon, 1 to 30."),
			"max_results": integerSchema("Maximum anomaly or segment rows, 1 to 20."),
			"test":        stringSchema("Optional hypothesis test selector, e.g. welch or chi_square_2x2."),
		}),
	}
}

func objectSchema(description string, properties map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":        "object",
		"description": description,
		"properties":  properties,
	}
}
