package assistant

import (
	"errors"
	"strings"
	"testing"

	"TaipeiCityDashboardBE/app/models"
)

func TestFilterComponentsResolvesRequestedCityVariant(t *testing.T) {
	items := []models.CityComponentScore{{
		ID: 215, Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構",
		City: "taipei", Score: 0.8484,
	}}
	filtered := filterComponentsWithResolver(items, "metrotaipei", 5,
		func(id int, city string) (models.CityComponent, error) {
			if id != 215 || city != "metrotaipei" {
				return models.CityComponent{}, errors.New("unexpected component")
			}
			return models.CityComponent{
				ID: 215, Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構",
				City: "metrotaipei", QueryType: "three_d",
			}, nil
		})

	if len(filtered.related) != 1 {
		t.Fatalf("related length = %d, want 1", len(filtered.related))
	}
	if filtered.related[0].City != "metrotaipei" || filtered.related[0].QueryType != "three_d" {
		t.Fatalf("related[0] = %+v, want metrotaipei three_d", filtered.related[0])
	}
	if filtered.related[0].Score != 0.8484 {
		t.Fatalf("score = %v, want qdrant score", filtered.related[0].Score)
	}
	if filtered.cityResolutionDropCount != 0 {
		t.Fatalf("drop count = %d, want 0", filtered.cityResolutionDropCount)
	}
}

func TestFilterComponentsReportsCrossCityResolutionDrop(t *testing.T) {
	items := []models.CityComponentScore{{
		ID: 215, Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構",
		City: "taipei", Score: 0.8484,
	}}
	filtered := filterComponentsWithResolver(items, "metrotaipei", 5,
		func(id int, city string) (models.CityComponent, error) {
			return models.CityComponent{}, errors.New("city variant not found")
		})

	if len(filtered.related) != 0 {
		t.Fatalf("related = %+v, want none when city variant is missing", filtered.related)
	}
	if filtered.cityResolutionDropCount != 1 {
		t.Fatalf("drop count = %d, want 1", filtered.cityResolutionDropCount)
	}
	if len(filtered.cityResolutionDropPreview) != 1 {
		t.Fatalf("drop preview length = %d, want 1", len(filtered.cityResolutionDropPreview))
	}
	if filtered.cityResolutionDropPreview[0].Index != "aging_workforce_trend" {
		t.Fatalf("drop preview = %+v, want aging_workforce_trend", filtered.cityResolutionDropPreview[0])
	}
	if filtered.cityResolutionDropPreview[0].Name != "高齡就業人口之年增結構" {
		t.Fatalf("drop preview = %+v, want preserved component name", filtered.cityResolutionDropPreview[0])
	}
}

func TestFilterComponentsCapsCrossCityResolutionPreview(t *testing.T) {
	items := []models.CityComponentScore{
		{ID: 1, Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構", City: "taipei", Score: 0.8484},
		{ID: 2, Index: "metro_flow", Name: "捷運分時人流", City: "taipei", Score: 0.8421},
		{ID: 3, Index: "flood_alert", Name: "淹水告警分布", City: "taipei", Score: 0.8301},
	}
	filtered := filterComponentsWithResolver(items, "metrotaipei", 5,
		func(id int, city string) (models.CityComponent, error) {
			return models.CityComponent{}, errors.New("city variant not found")
		})

	if filtered.cityResolutionDropCount != 3 {
		t.Fatalf("drop count = %d, want 3", filtered.cityResolutionDropCount)
	}
	if len(filtered.cityResolutionDropPreview) != 2 {
		t.Fatalf("drop preview length = %d, want 2", len(filtered.cityResolutionDropPreview))
	}
	if filtered.cityResolutionDropPreview[0].Index != "aging_workforce_trend" || filtered.cityResolutionDropPreview[1].Index != "metro_flow" {
		t.Fatalf("drop preview = %+v, want first two dropped indexes", filtered.cityResolutionDropPreview)
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

func TestSearchComponentsEnvelopeSuppressesLowConfidenceRecommendations(t *testing.T) {
	envelope := searchComponentsEnvelope("防災", []RelatedComponent{
		{ID: 1, Index: "metro_flow", Name: "捷運分時人流", City: "metrotaipei", Score: 0.8010},
		{ID: 2, Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構", City: "metrotaipei", Score: 0.7990},
	}, []string{"使用 Qdrant 語意檢索公開儀表板組件 metadata。"}, map[string]interface{}{
		"query":  "防災",
		"status": "ok",
	})

	if len(envelope.RelatedComponents) != 0 {
		t.Fatalf("related components = %+v, want none for low-confidence query", envelope.RelatedComponents)
	}
	if !containsConfidenceNote(envelope.ConfidenceNotes, "暫不直接推薦組件") {
		t.Fatalf("confidence notes = %+v, want low-confidence suppression note", envelope.ConfidenceNotes)
	}
}

func TestSearchComponentsEnvelopeSuppressesWeakFallbackRecommendations(t *testing.T) {
	envelope := searchComponentsEnvelope("通勤", []RelatedComponent{
		{ID: 3, Index: "metro_flow", Name: "捷運分時人流", City: "metrotaipei", Score: 0.2},
	}, []string{"語意索引暫不可用，已改用既有 dashboard metadata 關鍵字比對。"}, map[string]interface{}{
		"query":    "通勤",
		"status":   "degraded",
		"fallback": "db_metadata",
	})

	if len(envelope.RelatedComponents) != 0 {
		t.Fatalf("related components = %+v, want none for weak fallback query", envelope.RelatedComponents)
	}
	if !containsConfidenceNote(envelope.ConfidenceNotes, "請補充更具體的指標、城市或資料主題") {
		t.Fatalf("confidence notes = %+v, want refinement guidance", envelope.ConfidenceNotes)
	}
}

func TestSearchComponentsEnvelopeKeepsExactHighConfidenceRecommendation(t *testing.T) {
	envelope := searchComponentsEnvelope("aging_workforce_trend", []RelatedComponent{
		{ID: 214, Index: "dependency_aging", Name: "扶養比及老化指數", City: "metrotaipei", Score: 0.8515},
		{ID: 215, Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構", City: "metrotaipei", Score: 0.8484},
	}, []string{"使用 Qdrant 語意檢索公開儀表板組件 metadata。"}, map[string]interface{}{
		"query":  "aging_workforce_trend",
		"status": "ok",
	})

	if len(envelope.RelatedComponents) == 0 {
		t.Fatal("expected exact-query recommendation to remain available")
	}
	if envelope.RelatedComponents[0].Index != "aging_workforce_trend" {
		t.Fatalf("first related index = %s, want aging_workforce_trend", envelope.RelatedComponents[0].Index)
	}
}

func TestSearchComponentsEnvelopeAddsCityResolutionDropNote(t *testing.T) {
	envelope := searchComponentsEnvelope("commute_flow", []RelatedComponent{
		{ID: 3, Index: "commute_flow", Name: "通勤人口熱區", City: "metrotaipei", Score: 0.88},
	}, []string{"使用 Qdrant 語意檢索公開儀表板組件 metadata。"}, map[string]interface{}{
		"query":                      "commute_flow",
		"status":                     "ok",
		"city_resolution_drop_count": 1,
		"city_resolution_drop_preview": []cityResolutionDropPreview{{
			Index: "aging_workforce_trend", Name: "高齡就業人口之年增結構", City: "taipei",
		}},
	})

	if !containsConfidenceNote(envelope.ConfidenceNotes, "其他城市有對應組件") {
		t.Fatalf("confidence notes = %+v, want cross-city resolution note", envelope.ConfidenceNotes)
	}
	if !containsConfidenceNote(envelope.ConfidenceNotes, "這不代表沒有相近主題") {
		t.Fatalf("confidence notes = %+v, want explicit non-empty-result clarification", envelope.ConfidenceNotes)
	}
	if !containsConfidenceNote(envelope.ConfidenceNotes, "aging_workforce_trend") {
		t.Fatalf("confidence notes = %+v, want dropped component preview", envelope.ConfidenceNotes)
	}
	if got, ok := envelope.Data.(map[string]interface{}); !ok || got["city_resolution_drop_count"] != 1 {
		t.Fatalf("data = %#v, want city_resolution_drop_count=1", envelope.Data)
	}
}

func containsConfidenceNote(notes []string, want string) bool {
	for _, note := range notes {
		if strings.Contains(note, want) {
			return true
		}
	}
	return false
}
