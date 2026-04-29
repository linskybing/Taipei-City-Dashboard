package assistant

import (
	"math"
	"testing"
)

func TestDescribePointsComputesSummary(t *testing.T) {
	points := []statPoint{{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}}
	summary := describePoints(points)
	if summary.Count != 4 || summary.Min != 1 || summary.Max != 4 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if summary.Median != 2.5 {
		t.Fatalf("Median = %v, want 2.5", summary.Median)
	}
}

func TestCalculateTrendIncreasing(t *testing.T) {
	points := []statPoint{{Value: 1}, {Value: 3}, {Value: 5}, {Value: 7}}
	trend := calculateTrend(points)
	if trend.Direction != "increasing" || trend.Strength != "high" {
		t.Fatalf("unexpected trend: %#v", trend)
	}
	if trend.Slope != 2 {
		t.Fatalf("Slope = %v, want 2", trend.Slope)
	}
}

func TestBuildSeasonalCardRejectsShortSeries(t *testing.T) {
	ds := statDataset{ComponentID: 1, City: "taipei", QueryType: "time"}
	points := []statPoint{{Value: 1}, {Value: 2}, {Value: 3}}
	_, err := buildSeasonalCard(ds, points, statToolInput{Options: map[string]interface{}{"period": 7.0}})
	if err == nil {
		t.Fatal("expected short seasonal series error")
	}
}

func TestBuildHypothesisCardWelch(t *testing.T) {
	ds := statDataset{ComponentID: 1, City: "taipei", QueryType: "time"}
	points := []statPoint{
		{Series: "a", Value: 10}, {Series: "a", Value: 12}, {Series: "a", Value: 14},
		{Series: "b", Value: 2}, {Series: "b", Value: 3}, {Series: "b", Value: 4},
	}
	card, err := buildHypothesisCard(ds, points, statToolInput{Options: map[string]interface{}{}})
	if err != nil {
		t.Fatalf("buildHypothesisCard returned error: %v", err)
	}
	data := card.Data.(map[string]interface{})
	if math.Abs(data["difference"].(float64)-9) > 0.001 {
		t.Fatalf("difference = %#v, want about 9", data["difference"])
	}
}

func TestForecastSeriesUsesSeasonalNaiveWhenPossible(t *testing.T) {
	points := []statPoint{
		{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4},
		{Value: 10}, {Value: 20}, {Value: 30}, {Value: 40},
	}
	forecast, method := forecastSeries(points, 2, 4)
	if method != "seasonal_naive" {
		t.Fatalf("method = %s, want seasonal_naive", method)
	}
	if forecast[0].Y != 10 || forecast[1].Y != 20 {
		t.Fatalf("unexpected forecast: %#v", forecast)
	}
}
