package assistant

func componentSearchGuardrails() []string {
	return []string{
		"只根據回傳的 related_components 與公開資料來源建議組件；未回傳的組件不得假定存在。",
		"若需要單一組件細節或 chart sample，下一步呼叫 get_component_snapshot。",
		"跨臺北/雙北比較必須呼叫 compare_city_components，不以名稱相似自行推論。",
	}
}
