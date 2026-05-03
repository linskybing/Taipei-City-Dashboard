import { DIFFICULTY_SCORE_EXPRESSION } from "./parkingDifficulty";

const LOCAL_PARKING_COMPONENTS = new Set([
	"parking_supply_metrotaipei",
	"parking_price_metrotaipei",
	"parking_difficulty_spectrum_metrotaipei",
]);

const PARKING_SUPPLY_INDEX = "parking_supply_metrotaipei";
const PARKING_PRICE_INDEX = "parking_price_metrotaipei";
const PARKING_DIFFICULTY_INDEX = "parking_difficulty_spectrum_metrotaipei";
const POINT_GEOJSON_BY_CITY = {
	taipei: "parking_supply_points_taipei",
	metrotaipei: "parking_supply_points_metrotaipei",
};
const PARKING_GEOJSON_BY_CITY = {
	taipei: "parking_public_points_taipei",
	metrotaipei: "parking_public_points_metrotaipei",
};
const ONSTREET_GEOJSON_BY_CITY = {
	taipei: "parking_onstreet_points_taipei",
	metrotaipei: "parking_onstreet_points_metrotaipei",
};
const DIFFICULTY_GEOJSON_BY_CITY = {
	taipei: "parking_difficulty_points_taipei_wanhua",
	metrotaipei: "parking_difficulty_points_metrotaipei_wanhua",
};
const AVAILABLE_RATE = [
	"coalesce",
	["to-number", ["get", "available_rate"]],
	-1,
];
const HAS_LOW_REMAINING_SPACE = [
	"all",
	[">", AVAILABLE_RATE, 0],
	["<", AVAILABLE_RATE, 0.35],
];
const PARKING_STATUS_COLORS = [
	"case",
	HAS_LOW_REMAINING_SPACE,
	"#f97316",
	["==", ["get", "status_label"], "空位"],
	"#16a34a",
	"#6b7280",
];
const CAPACITY = ["coalesce", ["to-number", ["get", "capacity_total"]], 0];
const PARKING_SUPPLY_CIRCLE_RADII = {
	"停車場":
		[
			"interpolate",
			["linear"],
			["log10", ["+", CAPACITY, 1]],
			0,
			3,
			1.7,
			4.6,
			2.18,
			5.4,
			2.6,
			6.7,
			3,
			8,
		],
	"路邊停車格":
		[
			"interpolate",
			["linear"],
			["zoom"],
			10,
			6,
			12,
			8,
			14,
			10.5,
			16,
			13,
		],
};
const PARKING_PRICE_CIRCLE_RADIUS = ["interpolate", ["linear"], ["log10", ["+", ["coalesce", ["to-number", ["get", "price_value"]], 0], 1]], 0, 3, 1.3, 5, 1.6, 8, 1.9, 14];
const DIFFICULTY_SCORE = ["max", 0, ["min", 100, DIFFICULTY_SCORE_EXPRESSION]];
const DIFFICULTY_COLOR = [
	"interpolate",
	["linear"],
	DIFFICULTY_SCORE,
	20,
	"#22c55e",
	45,
	"#facc15",
	65,
	"#f97316",
	85,
	"#dc2626",
];
const DIFFICULTY_RADIUS = [
	"interpolate",
	["linear"],
	["zoom"],
	10,
	["interpolate", ["linear"], DIFFICULTY_SCORE, 0, 3, 100, 5],
	16,
	["interpolate", ["linear"], DIFFICULTY_SCORE, 0, 7, 100, 13],
];

function getParkingMapIndex(component, config) {
	if (component.index === PARKING_DIFFICULTY_INDEX) {
		return (
			DIFFICULTY_GEOJSON_BY_CITY[component.city] ||
			DIFFICULTY_GEOJSON_BY_CITY.metrotaipei
		);
	}
	if (component.index === PARKING_SUPPLY_INDEX && config?.title === "停車場") {
		return (
			PARKING_GEOJSON_BY_CITY[component.city] ||
			PARKING_GEOJSON_BY_CITY.metrotaipei
		);
	}
	if (
		component.index === PARKING_SUPPLY_INDEX &&
		config?.title === "路邊停車格"
	) {
		return (
			ONSTREET_GEOJSON_BY_CITY[component.city] ||
			ONSTREET_GEOJSON_BY_CITY.metrotaipei
		);
	}
	return (
		POINT_GEOJSON_BY_CITY[component.city] ||
		POINT_GEOJSON_BY_CITY.metrotaipei
	);
}

function getParkingPaint(component, config) {
	const paint = { ...(config.paint || {}) };
	delete paint["layer-default-visibility"];

	if (component.index === PARKING_SUPPLY_INDEX) {
		const circleRadius = PARKING_SUPPLY_CIRCLE_RADII[config.title];
		if (circleRadius) {
			paint["circle-radius"] = circleRadius;
		}
		if (config.title === "停車場") {
			paint["circle-color"] = PARKING_STATUS_COLORS;
			paint["circle-stroke-color"] = "rgba(0,0,0,0)";
			paint["circle-stroke-width"] = 0;
		}
		if (config.title === "路邊停車格") {
			paint["circle-color"] = PARKING_STATUS_COLORS;
			paint["circle-stroke-color"] = "#ffffff";
			paint["circle-stroke-width"] = 1.2;
			paint["circle-opacity"] = 0.9;
		}
		paint["layer-default-visibility"] = "none";
	}
	if (component.index === PARKING_PRICE_INDEX && config.title === "價格熱區") {
		paint["circle-radius"] = PARKING_PRICE_CIRCLE_RADIUS;
		paint["layer-default-visibility"] = "none";
	}
	if (component.index === PARKING_DIFFICULTY_INDEX) {
		paint["circle-color"] = DIFFICULTY_COLOR;
		paint["circle-radius"] = DIFFICULTY_RADIUS;
		paint["circle-stroke-color"] = "#ffffff";
		paint["circle-stroke-width"] = 1;
		paint["circle-opacity"] = 0.88;
	}

	return paint;
}

function getParkingFilter(component, config) {
	if (component.index !== PARKING_SUPPLY_INDEX) {
		return config.filter;
	}
	return null;
}

function normalizeLocalParkingLayerConfig(component, config, geojsonIndex) {
	return {
		...config,
		default_on: false,
		source: "geojson",
		index: geojsonIndex || getParkingMapIndex(component, config),
		filter: getParkingFilter(component, config),
		paint: getParkingPaint(component, config),
	};
}

export function normalizeLocalParkingMapConfig(component) {
	if (
		!component ||
		!LOCAL_PARKING_COMPONENTS.has(component.index) ||
		!Array.isArray(component.map_config)
	) {
		return component;
	}
	return {
		...component,
		map_config: component.map_config
			.filter((config) => config.type !== "fill")
			.map((config) =>
				normalizeLocalParkingLayerConfig(component, config),
			),
	};
}
