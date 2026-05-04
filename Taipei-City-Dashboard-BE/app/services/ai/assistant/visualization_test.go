package assistant

import "testing"

func TestBuildVisualizationRefsUsesComponentReferences(t *testing.T) {
	extras := ResponseExtras{
		RelatedComponents: []RelatedComponent{{
			Index:     "metro_station_hourly_flow",
			Name:      "捷運分時人流",
			City:      "metrotaipei",
			QueryType: "time",
		}},
		AnalysisCards: []AnalysisCard{{
			Tool:            "trend_detect",
			Headline:        "尖峰人流呈現上升",
			Assumptions:     []string{"以既有 component time data 判讀。"},
			ConfidenceLabel: "medium",
			Uncertainty: AnalysisUncertainty{
				Interval:         "not_applicable",
				DataQualityScore: 0.8,
				ConfidenceLabel:  "medium",
				CalibrationNote:  "描述性趨勢，不代表因果。",
			},
		}},
	}

	refs := BuildVisualizationRefs(extras, "ai_chatlog:42")
	if len(refs) != 1 {
		t.Fatalf("len(refs) = %d, want 1", len(refs))
	}
	ref := refs[0]
	if ref.ComponentIndex != "metro_station_hourly_flow" {
		t.Fatalf("ComponentIndex = %q", ref.ComponentIndex)
	}
	if ref.QueryType != "time" || ref.City != "metrotaipei" {
		t.Fatalf("unexpected query/city: %+v", ref)
	}
	if ref.Summary.Headline != "尖峰人流呈現上升" {
		t.Fatalf("headline = %q", ref.Summary.Headline)
	}
	if ref.AuditRef != "ai_chatlog:42" || ref.TraceID == "" {
		t.Fatalf("missing audit trace: %+v", ref)
	}
}
