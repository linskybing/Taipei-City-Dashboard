export function buildAnalysisChartXAxis(type, series) {
	const labels = { style: { colors: "#aeb9c6" }, hideOverlappingLabels: true, trim: true };
	if (!usesDatetimeAxis(type, series)) return { labels };
	return {
		type: "datetime",
		tickAmount: Math.min(maxPointCount(series), 5),
		labels: {
			...labels,
			trim: false,
			datetimeUTC: false,
			formatter: (_value, timestamp) => formatDateLabel(timestamp),
		},
	};
}

function usesDatetimeAxis(type, series) {
	if (type !== "line" && type !== "rangeArea") return false;
	const points = series.flatMap((item) => (Array.isArray(item.data) ? item.data : []));
	return points.length >= 2 && points.every((point) => isIsoDateLike(point?.x));
}

function maxPointCount(series) {
	return series.reduce((max, item) => Math.max(max, Array.isArray(item.data) ? item.data.length : 0), 0);
}

function isIsoDateLike(value) {
	return typeof value === "string"
		&& /^\d{4}[-/]\d{2}(?:[-/]\d{2})?(?:[ T]\d{2}:\d{2}(?::\d{2})?(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)?$/.test(value.trim())
		&& !Number.isNaN(Date.parse(value));
}

function formatDateLabel(timestamp) {
	const date = new Date(timestamp);
	if (Number.isNaN(date.getTime())) return "-";
	const month = `${date.getMonth() + 1}`.padStart(2, "0");
	const day = `${date.getDate()}`.padStart(2, "0");
	return `${date.getFullYear()}/${month}/${day}`;
}