package assistant

import "testing"

func TestStatCardsAttachSemanticVisualizationKinds(t *testing.T) {
	ds := statDataset{ComponentID: 1, City: "taipei", QueryType: "time"}
	points := []statPoint{
		{Series: "value", X: "1", Value: 1}, {Series: "value", X: "2", Value: 2},
		{Series: "value", X: "3", Value: 3}, {Series: "value", X: "4", Value: 4},
		{Series: "value", X: "5", Value: 100}, {Series: "value", X: "6", Value: 6},
		{Series: "value", X: "7", Value: 7}, {Series: "value", X: "8", Value: 8},
	}

	cases := []struct {
		name string
		card AnalysisCard
		kind string
	}{
		{"clean", mustCard(buildCleanCard(ds, points, statToolInput{})), "quality_summary"},
		{"descriptive", mustCard(buildDescriptiveCard(ds, points, statToolInput{})), "box_plot"},
		{"trend", mustCard(buildTrendCard(ds, points, statToolInput{})), "trend_line"},
		{"seasonal", mustCard(buildSeasonalCard(ds, points, statToolInput{Options: map[string]interface{}{"period": 4.0}})), "seasonal_profile"},
		{"anomaly", mustCard(buildAnomalyCard(ds, points, statToolInput{})), "anomaly_context"},
		{"forecast", mustCard(buildForecastCard(ds, points, statToolInput{})), "forecast_band"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if kind := visualizationKind(tc.card); kind != tc.kind {
				t.Fatalf("visualization kind = %q, want %q", kind, tc.kind)
			}
		})
	}
}

func TestHypothesisVisualizations(t *testing.T) {
	ds := statDataset{ComponentID: 1, City: "taipei", QueryType: "three_d"}
	welchPoints := []statPoint{
		{Series: "a", X: "1", Value: 10}, {Series: "a", X: "2", Value: 12},
		{Series: "b", X: "1", Value: 3}, {Series: "b", X: "2", Value: 4},
	}
	welch := mustCard(buildHypothesisCard(ds, welchPoints, statToolInput{}))
	if kind := visualizationKind(welch); kind != "effect_interval" {
		t.Fatalf("welch kind = %q, want effect_interval", kind)
	}

	tablePoints := []statPoint{
		{Series: "a", X: "yes", Value: 10}, {Series: "a", X: "no", Value: 2},
		{Series: "b", X: "yes", Value: 3}, {Series: "b", X: "no", Value: 9},
	}
	input := statToolInput{Options: map[string]interface{}{"test": "fisher_2x2"}}
	table := mustCard(buildHypothesisCard(ds, tablePoints, input))
	if kind := visualizationKind(table); kind != "contingency_heatmap" {
		t.Fatalf("2x2 kind = %q, want contingency_heatmap", kind)
	}
}

func TestVisualizationBoundsTrendPoints(t *testing.T) {
	points := make([]statPoint, 40)
	for i := range points {
		points[i] = statPoint{Series: "value", X: string(rune('A' + i%26)), Value: float64(i)}
	}
	viz := trendVisualization(points, calculateTrend(points))
	series := viz["series"].([]map[string]interface{})
	observed := series[0]["data"].([]map[string]interface{})
	if len(observed) != maxVisualizationPoints {
		t.Fatalf("observed points = %d, want %d", len(observed), maxVisualizationPoints)
	}
}

func visualizationKind(card AnalysisCard) string {
	data := card.Data.(map[string]interface{})
	viz := data["visualization"].(map[string]interface{})
	return viz["kind"].(string)
}

func mustCard(card AnalysisCard, err error) AnalysisCard {
	if err != nil {
		panic(err)
	}
	return card
}
