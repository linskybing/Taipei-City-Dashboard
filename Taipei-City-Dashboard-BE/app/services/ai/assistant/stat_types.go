package assistant

import "time"

type statToolInput struct {
	ComponentID int                    `json:"component_id"`
	City        string                 `json:"city"`
	TimeRange   statTimeRange          `json:"time_range"`
	SeriesName  string                 `json:"series_name"`
	Options     map[string]interface{} `json:"options"`
}

type statTimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type statDataset struct {
	ComponentID       int                `json:"component_id"`
	City              string             `json:"city"`
	QueryType         string             `json:"query_type"`
	Points            []statPoint        `json:"-"`
	InvalidPoints     int                `json:"invalid_points"`
	Sources           []Source           `json:"sources"`
	RelatedComponents []RelatedComponent `json:"related_components"`
}

type statPoint struct {
	Series  string    `json:"series"`
	X       string    `json:"x"`
	Time    time.Time `json:"-"`
	HasTime bool      `json:"-"`
	Value   float64   `json:"value"`
}

type statSummary struct {
	Count  int     `json:"count"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	StdDev float64 `json:"stddev"`
	Q1     float64 `json:"q1"`
	Q3     float64 `json:"q3"`
}

type trendResult struct {
	Direction     string  `json:"direction"`
	Strength      string  `json:"strength"`
	Slope         float64 `json:"slope"`
	Intercept     float64 `json:"intercept"`
	R             float64 `json:"r"`
	R2            float64 `json:"r2"`
	PercentChange float64 `json:"percent_change"`
}

type forecastPoint struct {
	X     string  `json:"x"`
	Y     float64 `json:"y"`
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
}
