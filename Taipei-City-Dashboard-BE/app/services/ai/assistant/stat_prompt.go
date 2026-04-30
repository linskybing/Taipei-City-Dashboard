package assistant

func statRoutingInstruction() string {
	return `統計工具路由：
- 若使用者要求「完整統計診斷」、「完整分析」、「統計工具全跑」或「所有統計」，且已有 component_id，同一輪可並行呼叫 clean_impute、descriptive_report、trend_detect、seasonal_decompose、anomaly_detect、hypothesis_test、forecast_short_mid；工具回傳 unavailable 時仍要列入資料信心，不得改寫成成功。
- 若使用者要求統計分析但沒有 component_id，先用 search_components 取得候選組件，再用 related_components 的 id 作為 component_id；若候選不明確，先請使用者指定組件。
- 資料品質、缺漏、重複、gap、補值建議：使用 clean_impute。
- 平均、中位數、四分位、標準差、最大最小、排名或區段摘要：使用 descriptive_report。
- 上升下降、斜率、百分比變化、趨勢強度：使用 trend_detect。
- 週期、季節性、尖離峰、週期 bucket 或殘差偏離：使用 seasonal_decompose。
- 異常、離群、暴增暴跌、警戒點排序：使用 anomaly_detect。
- A/B、兩組差異、政策前後比較、2x2 類別表：使用 hypothesis_test；資料形狀不合時回報 unavailable 與假設限制。
- 短中期預測、未來幾期、baseline forecast、區間：使用 forecast_short_mid，並標示「機制未變」假設。`
}
