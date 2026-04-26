package assistant

import (
	"TaipeiCityDashboardBE/app/models"
	"time"

	"github.com/lib/pq"
)

type ChartSample struct {
	QueryType  string      `json:"query_type"`
	Data       interface{} `json:"data,omitempty"`
	Categories []string    `json:"categories,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func getComponent(componentID int, city string) (models.CityComponent, error) {
	return models.GetComponentByID(componentID, city)
}

func getComponentIndex(componentID int) (string, error) {
	components, err := models.GetComponentByIDAll(componentID)
	if err != nil {
		return "", err
	}
	if len(components) == 0 {
		return "", nil
	}
	return components[0].Index, nil
}

func getComponentID(index string) (int, error) {
	var component models.Component
	err := models.DBManager.
		Table("components").
		Where("index = ?", index).
		First(&component).
		Error
	return int(component.ID), err
}

func getDashboardComponents(index string, city string, limit int) ([]models.CityComponent, error) {
	type componentArray struct {
		Components pq.Int64Array `gorm:"type:int[]"`
	}
	var row componentArray
	err := models.DBManager.
		Table("dashboards").
		Select("components").
		Where("index = ?", index).
		First(&row).
		Error
	if err != nil {
		return nil, err
	}

	components := make([]models.CityComponent, 0, len(row.Components))
	for _, id := range row.Components {
		component, err := getComponent(int(id), city)
		if err == nil {
			components = append(components, component)
		}
		if len(components) >= limit {
			break
		}
	}
	return components, nil
}

func getChartSample(componentID int, city string) ChartSample {
	queryType, queryString, err := models.GetComponentChartDataQuery(componentID, city)
	sample := ChartSample{QueryType: queryType}
	if err != nil {
		sample.Error = err.Error()
		return sample
	}
	if queryType == "" || queryString == "" {
		sample.Error = "no chart query"
		return sample
	}

	timeFrom, timeTo := defaultTimeRange()
	switch queryType {
	case "two_d":
		data, err := models.GetTwoDimensionalData(&queryString, timeFrom, timeTo)
		sample.Data, sample.Error = trimTwoD(data), errorString(err)
	case "three_d", "percent":
		data, categories, err := models.GetThreeDimensionalData(&queryString, timeFrom, timeTo)
		sample.Data, sample.Categories, sample.Error = trimThreeD(data), trimStrings(categories, 5), errorString(err)
	case "time":
		data, err := models.GetTimeSeriesData(&queryString, timeFrom, timeTo)
		sample.Data, sample.Error = trimTimeSeries(data), errorString(err)
	case "map_legend":
		data, err := models.GetMapLegendData(&queryString, timeFrom, timeTo)
		sample.Data, sample.Error = trimMapLegend(data), errorString(err)
	default:
		sample.Error = "unsupported query type"
	}
	return sample
}

func defaultTimeRange() (string, string) {
	now := time.Now().In(time.FixedZone("Asia/Taipei", 8*60*60))
	return now.Add(-24 * time.Hour).Format("2006-01-02T15:04:05+08:00"),
		now.Format("2006-01-02T15:04:05+08:00")
}

func trimTwoD(data []models.TwoDimensionalDataOutput) []models.TwoDimensionalDataOutput {
	for i := range data {
		if len(data[i].Data) > 5 {
			data[i].Data = data[i].Data[:5]
		}
	}
	if len(data) > 3 {
		return data[:3]
	}
	return data
}

func trimThreeD(data []models.ThreeDimensionalDataOutput) []models.ThreeDimensionalDataOutput {
	for i := range data {
		if len(data[i].Data) > 5 {
			data[i].Data = data[i].Data[:5]
		}
	}
	if len(data) > 5 {
		return data[:5]
	}
	return data
}

func trimTimeSeries(data []models.TimeSeriesDataOutput) []models.TimeSeriesDataOutput {
	for i := range data {
		if len(data[i].Data) > 5 {
			data[i].Data = data[i].Data[len(data[i].Data)-5:]
		}
	}
	if len(data) > 3 {
		return data[:3]
	}
	return data
}

func trimMapLegend(data []models.MapLegendData) []models.MapLegendData {
	if len(data) > 10 {
		return data[:10]
	}
	return data
}

func trimStrings(values []string, limit int) []string {
	if len(values) > limit {
		return values[:limit]
	}
	return values
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
