package assistant

type DecisionPlaybook struct {
	Theme      string   `json:"theme"`
	Label      string   `json:"label"`
	Question   string   `json:"question"`
	MapLayer   string   `json:"map_layer"`
	TrendChart string   `json:"trend_chart"`
	RankChart  string   `json:"rank_chart"`
	Signals    []string `json:"signals"`
	Steps      []string `json:"steps"`
	Guardrails []string `json:"guardrails"`
}

func playbookFor(theme string) DecisionPlaybook {
	if playbook, ok := playbooks()[theme]; ok {
		return playbook
	}
	return playbooks()[DefaultTheme]
}

func playbooks() map[string]DecisionPlaybook {
	return map[string]DecisionPlaybook{
		"auto": makePlaybook("auto", "六大主題自動判斷", "使用者問題最接近哪個城市決策主題？",
			"先找可用地圖或空間組件", "找時間序列或近期變化", "找行政區/資源排行",
			[]string{"地點", "時間", "受影響對象", "可用組件"},
			[]string{"判定主題", "搜尋相關組件", "取得必要快照", "整理下一步"},
			[]string{"主題不明時先說明需要補充的地點、時間與對象"}),
		"commuting": makePlaybook("commuting", "通勤失敗預警", "哪個轉乘或最後一哩節點最可能失敗？",
			"轉乘風險圖層", "ETA/雨量/車位波動", "替代路徑穩定度",
			[]string{"公車 ETA", "YouBike 可用量", "雨量", "道路熱區"},
			[]string{"標記高風險節點", "比較替代運具", "輸出提前量或改道建議"},
			[]string{"不可宣稱即時預測超出資料更新頻率"}),
		"disaster": makePlaybook("disaster", "避難可達性推演", "未來數小時哪個避難點仍可到達且合適？",
			"避難可達性圖層", "雨量/水位/示警時間線", "優先避難點排行",
			[]string{"NCDR 示警", "雨量", "水位", "避難收容資訊"},
			[]string{"辨識警戒區", "比對避難資源", "分級行動順序"},
			[]string{"防災建議需保守並提示依官方指揮為準"}),
		"environment": makePlaybook("environment", "熱空污低曝露行動窗", "何時何地的熱與空污暴露最低？",
			"熱空污複合風險圖層", "未來 12 小時曝露窗口", "行政區資源缺口排行",
			[]string{"AQI", "高溫", "鄉鎮預報", "公園綠地"},
			[]string{"找低曝露時段", "比對可近綠地", "提示敏感族群行動"},
			[]string{"不可把環境風險解讀成醫療診斷"}),
		"health": makePlaybook("health", "食安事件到就醫行動", "近期食安事件是否影響家庭行動？",
			"食安事件與就醫資源地圖", "腹瀉/事件趨勢", "食品類別或區域排行",
			[]string{"食藥署公告", "就醫趨勢", "醫療院所", "公告日期"},
			[]string{"拆解事件範圍", "比對在地症狀趨勢", "給保守就醫與避險建議"},
			[]string{"只提供風險解讀與資源，不提供診斷"}),
		"labor": makePlaybook("labor", "可持續轉職路徑", "哪組訓練、就服與照顧資源走得完？",
			"可持續就業半徑圖", "報名/訓練時程", "可行組合排行",
			[]string{"就服據點", "職訓課程", "托育資源", "社福中心"},
			[]string{"估算可達半徑", "比對照顧支持", "排序可完成路徑"},
			[]string{"不得替代正式就業媒合或資格審查"}),
		"culture": makePlaybook("culture", "文化供需缺口地圖", "哪裡活動供給未接住新住民需求？",
			"文化共融缺口圖層", "活動時間分布", "每萬人活動密度排行",
			[]string{"新住民人口", "藝文活動", "活動地點", "可近性"},
			[]string{"比對人口與活動密度", "找交通與語言友善活動", "提示資源缺口"},
			[]string{"不可臆測族群需求，需以公開資料保守描述"}),
	}
}

func makePlaybook(
	theme, label, question, mapLayer, trendChart, rankChart string,
	signals, steps, guardrails []string,
) DecisionPlaybook {
	return DecisionPlaybook{
		Theme:      theme,
		Label:      label,
		Question:   question,
		MapLayer:   mapLayer,
		TrendChart: trendChart,
		RankChart:  rankChart,
		Signals:    signals,
		Steps:      steps,
		Guardrails: guardrails,
	}
}
