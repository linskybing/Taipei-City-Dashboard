package assistant

import "github.com/tmc/langchaingo/llms"

func statisticalToolDefinitions() []llms.Tool {
	props := statToolProperties()
	return []llms.Tool{
		defineTool("clean_impute", "Use for missingness, duplicates, gaps, data-quality profile, and read-only imputation recommendations.", props, []string{"component_id"}),
		defineTool("descriptive_report", "Use for mean, median, stddev, quartiles, min/max, distribution, and top/bottom segment summaries.", props, []string{"component_id"}),
		defineTool("trend_detect", "Use for rising/falling trend, slope, percent change, and trend strength or confidence labels.", props, []string{"component_id"}),
		defineTool("seasonal_decompose", "Use for seasonality, period buckets, peak/off-peak comparison, and residual-style deviations.", props, []string{"component_id"}),
		defineTool("anomaly_detect", "Use for outliers, spikes, drops, warning candidates, and bounded z-score/IQR anomaly ranking.", props, []string{"component_id"}),
		defineTool("hypothesis_test", "Use for two-group comparison, before/after checks, A/B questions, Welch test, or 2x2 categorical tests.", props, []string{"component_id"}),
		defineTool("forecast_short_mid", "Use for short or medium baseline forecasts, future periods, intervals, and mechanism-unchanged caveats.", props, []string{"component_id"}),
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
