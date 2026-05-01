import { buildAnalysisChartXAxis } from "./chatAnalysisChartAxis";

const colors = ["#65b7ff", "#8bd17c", "#f7c66b", "#f27d7d", "#b9a6ff"];

export function buildAnalysisChart(card) {
	const viz = card?.data?.visualization || legacyVisualization(card);
	if (!viz?.kind) return null;
	const builders = {
		quality_summary: qualitySummary,
		box_plot: boxPlot,
		trend_line: trendLine,
		seasonal_profile: seasonalProfile,
		anomaly_context: anomalyContext,
		anomaly_table: anomalyTable,
		effect_interval: effectInterval,
		contingency_heatmap: contingencyHeatmap,
		forecast_band: forecastBand,
	};
	return builders[viz.kind]?.(viz) || null;
}

function qualitySummary(viz) {
	const metrics = Array.isArray(viz.series) ? viz.series : [];
	if (!metrics.length) return null;
	return { mode: "summary", title: "資料品質摘要", metrics, notes: viz.annotations || [], caption: viz.caption };
}

function boxPlot(viz) {
	const data = Array.isArray(viz.series) ? viz.series.filter((item) => Array.isArray(item.y) && item.y.length === 5) : [];
	if (!data.length) return null;
	return chart("boxPlot", "分布摘要盒鬚圖", [{ data }], { caption: annotationText(viz) || viz.caption });
}

function trendLine(viz) {
	const series = lineSeries(viz.series);
	if (series.length < 2) return null;
	return chart("line", "觀測值與趨勢線", series, { caption: annotationText(viz) || viz.caption, markers: 4 });
}

function seasonalProfile(viz) {
	const series = lineSeries(viz.series);
	if (!series.length) return null;
	return chart("line", "季節 bucket profile", series, { caption: viz.caption, markers: 4 });
}

function anomalyContext(viz) {
	const series = lineSeries(viz.series, true);
	if (!series.length) return null;
	return chart("line", "異常點脈絡", series, { caption: viz.caption, markers: [0, 6], yAnnotations: fences(viz.annotations) });
}

function anomalyTable(viz) {
	const rows = Array.isArray(viz.rows) ? viz.rows : [];
	if (!rows.length) return null;
	return { mode: "table", title: "候選異常點", rows, columns: ["x", "series", "value", "z_score", "severity"], caption: viz.caption };
}

function effectInterval(viz) {
	const data = Array.isArray(viz.series) ? viz.series.filter((item) => Array.isArray(item.y) && item.y.length === 2) : [];
	if (!data.length) return null;
	return chart("rangeBar", "效果量信賴區間", [{ name: "區間", data }], {
		caption: annotationText(viz) || viz.caption,
		horizontal: true,
		xAnnotations: [{ x: 0, borderColor: "#f27d7d", label: { text: "零線" } }],
	});
}

function contingencyHeatmap(viz) {
	const series = Array.isArray(viz.series) ? viz.series : [];
	if (!series.length) return null;
	return chart("heatmap", "2x2 類別檢定", series, { caption: annotationText(viz) || viz.caption });
}

function forecastBand(viz) {
	const source = Array.isArray(viz.series) ? viz.series : [];
	const series = source.map((item) => ({
		name: item.name,
		type: item.name?.includes("區間") ? "rangeArea" : "line",
		data: item.data || [],
	}));
	if (series.filter((item) => item.data.length).length < 2) return null;
	return chart("rangeArea", "預測與 prediction band", series, { caption: annotationText(viz) || viz.caption, markers: 3 });
}

function chart(type, title, series, extra = {}) {
	const xaxis = buildAnalysisChartXAxis(type, series);
	return {
		mode: "chart",
		type,
		height: extra.height || (xaxis.type === "datetime" ? 220 : 190),
		series,
		caption: extra.caption,
		options: {
			chart: { toolbar: { show: false }, foreColor: "#d8e1eb" },
			colors,
			dataLabels: { enabled: false },
			grid: { borderColor: "#2e343b", strokeDashArray: 3 },
			legend: { show: series.length > 1, position: "top", horizontalAlign: "left" },
			markers: { size: extra.markers || 0 },
			plotOptions: { bar: { borderRadius: 3, horizontal: !!extra.horizontal }, boxPlot: { colors: { upper: colors[0], lower: colors[2] } } },
			stroke: { curve: "straight", width: series.map((item) => (item.type === "rangeArea" ? 0 : 2)) },
			title: { text: title, style: { color: "#f4f7fb", fontSize: "12px", fontWeight: 700 } },
			tooltip: { theme: "dark", y: { formatter: formatNumber } },
			xaxis,
			yaxis: { labels: { style: { colors: "#aeb9c6" }, formatter: type === "heatmap" ? label : formatNumber } },
			annotations: { yaxis: extra.yAnnotations || [], xaxis: extra.xAnnotations || [] },
		},
	};
}

function lineSeries(items, mixed = false) {
	if (!Array.isArray(items)) return [];
	return items
		.map((item) => ({
			name: item.name,
			type: mixed && item.name?.includes("異常") ? "scatter" : "line",
			data: Array.isArray(item.data) ? item.data.filter((row) => row?.x && toNumber(row.y) !== null) : [],
		}))
		.filter((item) => item.data.length);
}

function fences(annotations = {}) {
	const rows = [
		["平均", annotations.mean, colors[1]],
		["下界", annotations.lower_fence, colors[2]],
		["上界", annotations.upper_fence, colors[3]],
	];
	return rows.filter(([, value]) => toNumber(value) !== null).map(([text, y, borderColor]) => ({ y, borderColor, label: { text } }));
}

function annotationText(viz) {
	if (!viz?.annotations || Array.isArray(viz.annotations)) return "";
	return Object.entries(viz.annotations)
		.filter(([, value]) => value !== undefined && value !== null && value !== "")
		.map(([key, value]) => `${key}: ${formatNumber(value)}`)
		.join(" / ");
}

function legacyVisualization(card) {
	const data = card?.data || {};
	if (card?.tool === "clean_impute") {
		return { kind: "quality_summary", series: [
			{ label: "無效值", value: data.invalid_points },
			{ label: "重複", value: data.duplicates },
			{ label: "時間缺口", value: data.time_gaps },
		], annotations: data.recommendations };
	}
	if (card?.tool === "descriptive_report" && data.summary) {
		const s = data.summary;
		return { kind: "box_plot", series: [{ x: "資料分布", y: [s.min, s.q1, s.median, s.q3, s.max] }], annotations: { mean: s.mean, stddev: s.stddev, count: s.count } };
	}
	if (card?.tool === "anomaly_detect" && Array.isArray(data.anomalies)) {
		return { kind: "anomaly_table", rows: data.anomalies, caption: "舊資料缺少觀測序列，只能以表格列出候選異常。" };
	}
	if (card?.tool === "hypothesis_test" && Array.isArray(data.table)) {
		const names = Array.isArray(data.series) ? data.series : ["A", "B"];
		return { kind: "contingency_heatmap", series: data.table.map((row, index) => ({ name: names[index], data: row.map((y, i) => ({ x: `欄 ${i + 1}`, y })) })) };
	}
	if (card?.tool === "forecast_short_mid" && Array.isArray(data.forecast)) {
		return { kind: "forecast_band", series: [
			{ name: "預測", data: data.forecast.map((item) => ({ x: item.x, y: item.y })) },
			{ name: "預測區間", data: data.forecast.map((item) => ({ x: item.x, y: [item.lower, item.upper] })) },
		] };
	}
	return null;
}

function toNumber(value) {
	const number = Number(value);
	return Number.isFinite(number) ? number : null;
}

function formatNumber(value) {
	const number = toNumber(value);
	if (number === null) return label(value);
	return Math.abs(number) >= 1000 || Math.abs(number) < 0.01 ? number.toPrecision(4) : `${Math.round(number * 1000) / 1000}`;
}

function label(value) {
	return value === undefined || value === null || value === "" ? "-" : `${value}`;
}
