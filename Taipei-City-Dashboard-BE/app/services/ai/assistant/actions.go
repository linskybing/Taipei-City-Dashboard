package assistant

func actionsFor(theme string, audience string) []string {
	actions := map[string][]string{
		"commuting": {
			"比對公共運輸、YouBike 與道路熱區，優先標記尖峰轉乘瓶頸。",
			"將替代路徑與運具調度建議分為即時、日內與政策三個層級。",
		},
		"disaster": {
			"交叉檢查雨量、水位、歷史災點與避難資源，避免單一訊號誤判。",
			"將高風險區域分級並對應通報、資源調度與民眾提醒。",
		},
		"environment": {
			"追蹤污染、綠能與節能指標趨勢，標示政策成效與異常區域。",
			"優先用可查證開放資料支持低碳或污染改善建議。",
		},
		"health": {
			"將食品稽查、醫療資源與公共衛生訊號分開判讀，再整理共同風險。",
			"對民眾建議保持保守，不提供診斷或個別醫療判斷。",
		},
		"labor": {
			"比對就業結構、年齡趨勢與福利據點，找出資源配置落差。",
			"把政策建議拆成就業支持、弱勢照護與服務可近性三類。",
		},
		"culture": {
			"整合活動、觀光熱點與商圈資源，找出可串聯的文化參與路徑。",
			"以跨區域參與度與資源分布評估文化共融成效。",
		},
		"auto": {
			"先判定問題最接近的六大主題，再搜尋相關儀表板組件取得佐證。",
			"若主題不明，回覆時列出需要補充的地點、時間與政策對象。",
		},
	}
	result := append([]string{}, actions[theme]...)
	if audience == "public" {
		result = append(result, "以市民可理解的行動指引呈現，避免內部調度術語。")
	} else {
		result = append(result, "以政府決策格式呈現優先順序、資料限制與跨局處協作建議。")
	}
	return result
}
