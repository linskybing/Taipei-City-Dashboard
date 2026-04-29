package assistant

import (
	"TaipeiCityDashboardBE/app/models"
	"fmt"
	"math"
	"strings"
	"time"
)

const statTimeLayout = "2006-01-02T15:04:05+08:00"

func parseStatInput(args string) (statToolInput, error) {
	var input statToolInput
	if err := parseArgs(args, &input); err != nil {
		return input, err
	}
	if input.ComponentID <= 0 {
		return input, fmt.Errorf("component_id is required")
	}
	ctx, err := NewContext(DefaultTheme, input.City, DefaultAudience, "")
	if err != nil {
		return input, err
	}
	input.City = ctx.City
	input.SeriesName = strings.TrimSpace(input.SeriesName)
	if input.Options == nil {
		input.Options = map[string]interface{}{}
	}
	return input, nil
}

func loadStatDataset(input statToolInput) (statDataset, error) {
	from, to, err := normalizeStatRange(input.TimeRange)
	if err != nil {
		return statDataset{}, err
	}
	queryType, queryString, err := models.GetComponentChartDataQuery(input.ComponentID, input.City)
	if err != nil {
		return statDataset{}, err
	}
	if queryType == "" || queryString == "" {
		return statDataset{}, fmt.Errorf("component has no chart query")
	}
	component, err := getComponent(input.ComponentID, input.City)
	if err != nil {
		return statDataset{}, err
	}
	ds := statDataset{
		ComponentID:       input.ComponentID,
		City:              input.City,
		QueryType:         queryType,
		Sources:           sourcesFromComponent(component),
		RelatedComponents: []RelatedComponent{relatedFromComponent(component, 0)},
	}
	if err := fillStatPoints(&ds, queryString, from, to); err != nil {
		return statDataset{}, err
	}
	if len(ds.Points) == 0 {
		return statDataset{}, fmt.Errorf("component data has no numeric points")
	}
	return ds, nil
}

func fillStatPoints(ds *statDataset, queryString string, from string, to string) error {
	switch ds.QueryType {
	case "two_d":
		data, err := models.GetTwoDimensionalData(&queryString, from, to)
		if err != nil {
			return err
		}
		for _, series := range data {
			for _, item := range series.Data {
				ds.addPoint("value", item.Xaxis, item.Data)
			}
		}
	case "three_d", "percent":
		data, categories, err := models.GetThreeDimensionalData(&queryString, from, to)
		if err != nil {
			return err
		}
		for _, series := range data {
			for i, value := range series.Data {
				x := fmt.Sprintf("%d", i+1)
				if i < len(categories) {
					x = categories[i]
				}
				ds.addPoint(series.Name, x, float64(value))
			}
		}
	case "time":
		data, err := models.GetTimeSeriesData(&queryString, from, to)
		if err != nil {
			return err
		}
		for _, series := range data {
			for _, item := range series.Data {
				ds.addPoint(series.Name, item.X, item.Y)
			}
		}
	case "map_legend":
		data, err := models.GetMapLegendData(&queryString, from, to)
		if err != nil {
			return err
		}
		for _, item := range data {
			series := item.Type
			if series == "" {
				series = "map_legend"
			}
			ds.addPoint(series, item.Name, item.Value)
		}
	default:
		return fmt.Errorf("unsupported query_type %s", ds.QueryType)
	}
	return nil
}

func (ds *statDataset) addPoint(series string, x string, value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		ds.InvalidPoints++
		return
	}
	point := statPoint{Series: emptyAs(series, "value"), X: x, Value: value}
	if parsed, ok := parseStatPointTime(x); ok {
		point.Time = parsed
		point.HasTime = true
	}
	ds.Points = append(ds.Points, point)
}

func selectStatPoints(ds statDataset, seriesName string) ([]statPoint, error) {
	points := make([]statPoint, 0, len(ds.Points))
	for _, point := range ds.Points {
		if seriesName == "" || point.Series == seriesName {
			points = append(points, point)
		}
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("no numeric data for requested series")
	}
	return points, nil
}

func normalizeStatRange(r statTimeRange) (string, string, error) {
	from, to := defaultTimeRange()
	if strings.TrimSpace(r.Start) != "" {
		normalized, err := normalizeStatTime(r.Start, false)
		if err != nil {
			return "", "", err
		}
		from = normalized
	}
	if strings.TrimSpace(r.End) != "" {
		normalized, err := normalizeStatTime(r.End, true)
		if err != nil {
			return "", "", err
		}
		to = normalized
	}
	return from, to, nil
}

func normalizeStatTime(value string, endOfDay bool) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) == len("2006-01-02") {
		if endOfDay {
			value += "T23:59:59+08:00"
		} else {
			value += "T00:00:00+08:00"
		}
	}
	if _, err := time.Parse(statTimeLayout, value); err != nil {
		return "", fmt.Errorf("invalid time_range timestamp %q", value)
	}
	return value, nil
}

func parseStatPointTime(value string) (time.Time, bool) {
	parsed, err := time.Parse(statTimeLayout, value)
	return parsed, err == nil
}

func emptyAs(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
