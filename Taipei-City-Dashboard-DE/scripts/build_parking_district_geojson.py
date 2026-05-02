import json
from pathlib import Path


def _normalize_features(source_path, output_path):
    source = json.loads(source_path.read_text(encoding="utf-8"))
    features = []
    for feature in source.get("features", []):
        props = feature.get("properties", {})
        district = props.get("TNAME")
        city = props.get("PNAME")
        if not district:
            continue
        features.append(
            {
                "type": "Feature",
                "geometry": feature.get("geometry"),
                "properties": {
                    "district": district,
                    "city": city,
                    "name": f"{city}{district}" if city else district,
                },
            }
        )
    output_path.write_text(
        json.dumps(
            {"type": "FeatureCollection", "features": features},
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )


def main():
    root = Path(__file__).resolve().parents[2]
    map_dir = root / "Taipei-City-Dashboard-FE" / "public" / "mapData"
    _normalize_features(
        map_dir / "taipei_town.geojson",
        map_dir / "tourism_parking_district_taipei.geojson",
    )
    _normalize_features(
        map_dir / "metrotaipei_town.geojson",
        map_dir / "tourism_parking_district_metrotaipei.geojson",
    )


if __name__ == "__main__":
    main()
