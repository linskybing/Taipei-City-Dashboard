const AVAILABILITY_WEIGHT = 45;
const PRICE_WEIGHT = 20;
const ONSTREET_WEIGHT = 10;
const CAPACITY_RELIEF_WEIGHT = 10;
const FULL_STATUSES = new Set(["已停", "不可用"]);
const PRICE_TIER_FALLBACK = {
	premium: PRICE_WEIGHT,
	high: 15,
	medium: 8,
	low: 3,
	free: 0,
};

export const PARKING_DIFFICULTY_RADIUS_KM = 0.3;

function clamp(value, min, max) {
	return Math.max(min, Math.min(max, value));
}

function numeric(value, fallback = 0) {
	const number = Number(value);
	return Number.isFinite(number) ? number : fallback;
}

export function availabilityPenalty(statusLabel, occupancyRate) {
	if (statusLabel === "空位") {
		return clamp(numeric(occupancyRate, 0.4), 0, 1) * 30;
	}
	if (FULL_STATUSES.has(statusLabel)) return AVAILABILITY_WEIGHT;
	return 38;
}

export function pricePenalty(priceValue, priceTier) {
	const value = Number(priceValue);
	if (Number.isFinite(value)) {
		return clamp(value / 80, 0, 1) * PRICE_WEIGHT;
	}
	return PRICE_TIER_FALLBACK[priceTier] ?? 10;
}

export function constructionPenalty(nearbyConstructionCount) {
	return clamp(numeric(nearbyConstructionCount) / 5, 0, 1) * 15;
}

export function onstreetPenalty(facilityKind) {
	return facilityKind === "onstreet" ? ONSTREET_WEIGHT : 0;
}

export function capacityRelief(facilityKind, capacityTotal) {
	if (facilityKind !== "public_parking") return 0;
	const capacity = Math.max(0, numeric(capacityTotal));
	return (
		clamp(Math.log10(capacity + 1) / Math.log10(501), 0, 1) *
		CAPACITY_RELIEF_WEIGHT
	);
}

export function getParkingDifficultyScore(properties = {}) {
	const score =
		availabilityPenalty(properties.status_label, properties.occupancy_rate) +
		pricePenalty(properties.price_value, properties.price_tier) +
		constructionPenalty(properties.nearby_construction_count) +
		onstreetPenalty(properties.facility_kind) -
		capacityRelief(properties.facility_kind, properties.capacity_total);
	return Math.round(clamp(score, 0, 100));
}

export function distanceKmFromCoordinates(a, b) {
	const [lng1, lat1] = a;
	const [lng2, lat2] = b;
	const toRad = (degree) => (degree * Math.PI) / 180;
	const dLat = toRad(lat2 - lat1);
	const dLng = toRad(lng2 - lng1);
	const rLat1 = toRad(lat1);
	const rLat2 = toRad(lat2);
	const halfChord =
		Math.sin(dLat / 2) ** 2 +
		Math.cos(rLat1) * Math.cos(rLat2) * Math.sin(dLng / 2) ** 2;
	return (
		6371 *
		2 *
		Math.atan2(Math.sqrt(halfChord), Math.sqrt(1 - halfChord))
	);
}

export function countNearbyConstructionSites(
	parkingFeature,
	constructionFeatures,
) {
	const parkingCoords = parkingFeature?.geometry?.coordinates;
	if (!Array.isArray(parkingCoords)) return 0;
	return constructionFeatures.reduce((count, feature) => {
		const coords = feature?.geometry?.coordinates;
		if (!Array.isArray(coords)) return count;
		if (
			Math.abs(coords[0] - parkingCoords[0]) > 0.004 ||
			Math.abs(coords[1] - parkingCoords[1]) > 0.004
		) {
			return count;
		}
		return distanceKmFromCoordinates(parkingCoords, coords) <=
			PARKING_DIFFICULTY_RADIUS_KM
			? count + 1
			: count;
	}, 0);
}

const OCCUPANCY = [
	"max",
	0,
	["min", 1, ["coalesce", ["to-number", ["get", "occupancy_rate"]], 0.4]],
];
const PRICE_VALUE = ["coalesce", ["to-number", ["get", "price_value"]], -1];
const CONSTRUCTION_COUNT = [
	"coalesce",
	["to-number", ["get", "nearby_construction_count"]],
	0,
];
const CAPACITY = ["coalesce", ["to-number", ["get", "capacity_total"]], 0];

export const DIFFICULTY_SCORE_EXPRESSION = [
	"+",
	[
		"case",
		["==", ["get", "status_label"], "空位"],
		["*", OCCUPANCY, 30],
		[
			"any",
			["==", ["get", "status_label"], "已停"],
			["==", ["get", "status_label"], "不可用"],
		],
		AVAILABILITY_WEIGHT,
		38,
	],
	[
		"case",
		[">=", PRICE_VALUE, 0],
		["*", ["max", 0, ["min", 1, ["/", PRICE_VALUE, 80]]], PRICE_WEIGHT],
		[
			"match",
			["get", "price_tier"],
			"premium",
			20,
			"high",
			15,
			"medium",
			8,
			"low",
			3,
			"free",
			0,
			10,
		],
	],
	["*", ["max", 0, ["min", 1, ["/", CONSTRUCTION_COUNT, 5]]], 15],
	["case", ["==", ["get", "facility_kind"], "onstreet"], 10, 0],
	[
		"*",
		-CAPACITY_RELIEF_WEIGHT,
		[
			"case",
			["==", ["get", "facility_kind"], "public_parking"],
			[
				"max",
				0,
				[
					"min",
					1,
					["/", ["log10", ["+", ["max", 0, CAPACITY], 1]], ["log10", 501]],
				],
			],
			0,
		],
	],
];
