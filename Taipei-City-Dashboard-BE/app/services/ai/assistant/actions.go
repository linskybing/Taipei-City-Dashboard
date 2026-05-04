package assistant

func actionsFor(theme string, audience string) []string {
	playbook := playbookFor(theme)
	result := make([]string, 0, len(playbook.Steps)+3)
	result = append(result, playbook.Steps...)
	result = append(result, "以 "+playbook.MapLayer+"、"+playbook.TrendChart+"、"+playbook.RankChart+" 組成可重用儀表板模板。")
	if audience == "public" {
		result = append(result, "以市民可理解的行動指引呈現，避免內部調度術語。")
	} else {
		result = append(result, "以政府決策格式呈現優先順序、資料限制與跨局處協作建議。")
	}
	result = append(result, playbook.Guardrails...)
	return result
}
