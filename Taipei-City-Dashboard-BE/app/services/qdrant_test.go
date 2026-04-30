package services

import (
	"testing"

	"TaipeiCityDashboardBE/app/models"
)

func TestQdrantComponentPointIDIncludesCity(t *testing.T) {
	taipei := models.QuertChartAndConponentForQdrant{
		ID: 215, Index: "aging_workforce_trend", City: "taipei",
	}
	metro := models.QuertChartAndConponentForQdrant{
		ID: 215, Index: "aging_workforce_trend", City: "metrotaipei",
	}
	if qdrantComponentPointID(taipei) == qdrantComponentPointID(metro) {
		t.Fatal("city variants must not share a Qdrant point id")
	}
}
