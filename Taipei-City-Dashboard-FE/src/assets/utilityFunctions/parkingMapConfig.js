const LOCAL_PARKING_COMPONENTS = new Set([
	"parking_supply_metrotaipei",
	"parking_price_metrotaipei",
]);

export function normalizeLocalParkingMapConfig(component) {
	if (
		!component ||
		!LOCAL_PARKING_COMPONENTS.has(component.index) ||
		!Array.isArray(component.map_config)
	) {
		return component;
	}
	const geojsonIndex =
		component.city === "taipei"
			? "parking_supply_points_taipei"
			: "parking_supply_points_metrotaipei";
	return {
		...component,
		map_config: component.map_config.map((config) => ({
			...config,
			default_on: true,
			source: "geojson",
			index: geojsonIndex,
		})),
	};
}
