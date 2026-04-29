package assistant

import (
	"context"
	"fmt"
)

type statCardBuilder func(statDataset, []statPoint, statToolInput) (AnalysisCard, error)

func CleanImpute(ctx context.Context, args string) (string, error) {
	return runStatTool(ctx, args, "clean_impute", buildCleanCard)
}

func DescriptiveReport(ctx context.Context, args string) (string, error) {
	return runStatTool(ctx, args, "descriptive_report", buildDescriptiveCard)
}

func TrendDetect(ctx context.Context, args string) (string, error) {
	return runStatTool(ctx, args, "trend_detect", buildTrendCard)
}

func SeasonalDecompose(ctx context.Context, args string) (string, error) {
	return runStatTool(ctx, args, "seasonal_decompose", buildSeasonalCard)
}

func AnomalyDetect(ctx context.Context, args string) (string, error) {
	return runStatTool(ctx, args, "anomaly_detect", buildAnomalyCard)
}

func HypothesisTest(ctx context.Context, args string) (string, error) {
	return runStatTool(ctx, args, "hypothesis_test", buildHypothesisCard)
}

func ForecastShortMid(ctx context.Context, args string) (string, error) {
	return runStatTool(ctx, args, "forecast_short_mid", buildForecastCard)
}

func runStatTool(ctx context.Context, args string, name string, build statCardBuilder) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	input, err := parseStatInput(args)
	if err != nil {
		return "", err
	}
	ds, err := loadStatDataset(input)
	if err != nil {
		return unavailableTool(name, err)
	}
	points, err := selectStatPoints(ds, input.SeriesName)
	if err != nil {
		return unavailableTool(name, err)
	}
	card, err := build(ds, points, input)
	if err != nil {
		return unavailableTool(name, err)
	}
	return marshalTool(toolEnvelope{
		Tool:              name,
		Sources:           ds.Sources,
		RelatedComponents: ds.RelatedComponents,
		ConfidenceNotes: []string{
			fmt.Sprintf("%s 使用既有 component %d 的 %s 資料。", name, ds.ComponentID, ds.QueryType),
		},
		AnalysisCards: []AnalysisCard{card},
		Data: map[string]interface{}{
			"component_id": ds.ComponentID,
			"city":         ds.City,
			"query_type":   ds.QueryType,
			"point_count":  len(points),
		},
	})
}
