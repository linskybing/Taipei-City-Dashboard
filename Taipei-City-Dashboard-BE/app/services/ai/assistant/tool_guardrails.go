package assistant

func componentSearchGuardrails() []string {
	return []string{
		"只根據回傳的 related_components 與公開資料來源建議組件；未回傳的組件不得假定存在。",
		"若 search_components 顯示 low confidence 或 degraded，只能把結果當線索，先要求補充更具體的指標、城市或資料主題。",
		"若需要單一組件細節或 chart sample，下一步呼叫 get_component_snapshot。",
		"跨臺北/雙北比較必須呼叫 compare_city_components，不以名稱相似自行推論。",
	}
}
