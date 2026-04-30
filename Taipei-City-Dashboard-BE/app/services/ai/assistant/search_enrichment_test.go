package assistant

import (
	"errors"
	"testing"

	"TaipeiCityDashboardBE/app/models"
)

func TestFilterComponentsResolvesRequestedCityVariant(t *testing.T) {
	items := []models.CityComponentScore{{
		ID: 215, Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構",
		City: "taipei", Score: 0.8484,
	}}
	related := filterComponentsWithResolver(items, "metrotaipei", 5,
		func(id int, city string) (models.CityComponent, error) {
			if id != 215 || city != "metrotaipei" {
				return models.CityComponent{}, errors.New("unexpected component")
			}
			return models.CityComponent{
				ID: 215, Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構",
				City: "metrotaipei", QueryType: "three_d",
			}, nil
		})

	if len(related) != 1 {
		t.Fatalf("related length = %d, want 1", len(related))
	}
	if related[0].City != "metrotaipei" || related[0].QueryType != "three_d" {
		t.Fatalf("related[0] = %+v, want metrotaipei three_d", related[0])
	}
	if related[0].Score != 0.8484 {
		t.Fatalf("score = %v, want qdrant score", related[0].Score)
	}
}

func TestRequestsStatAnalysisRecognizesChineseAnalysisQuery(t *testing.T) {
	if !requestsStatAnalysis("幫我分析高齡就業人口") {
		t.Fatal("expected analysis intent")
	}
	if requestsStatAnalysis("推薦相關城市儀表板") {
		t.Fatal("did not expect analysis intent")
	}
}

func TestRankRelatedComponentsBoostsExactQueryMatch(t *testing.T) {
	related := []RelatedComponent{
		{ID: 214, Index: "dependency_aging", Name: "扶養比及老化指數", City: "metrotaipei", Score: 0.8515},
		{ID: 215, Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構", City: "metrotaipei", Score: 0.8484},
	}
	ranked := rankRelatedComponents("高齡就業人口", related)
	if ranked[0].Index != "aging_workforce_trend" {
		t.Fatalf("first ranked index = %s, want aging_workforce_trend", ranked[0].Index)
	}
}
