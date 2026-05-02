import json
from pathlib import Path
import sys

import pandas as pd

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from dags.utils.offstreet_parking import (
    fetch_new_taipei_onstreet_parking,
    fetch_taipei_onstreet_parking,
)


EXPORT_COLUMNS = [
    "parking_id",
    "city",
    "district",
    "name",
    "address",
    "total_car",
    "available_car",
    "occupied_car",
    "occupancy_rate",
    "status_label",
    "data_time",
]


def _to_features(data):
    export = data[EXPORT_COLUMNS + ["lng", "lat"]].copy()
    export["data_time"] = export["data_time"].astype(str)
    features = []
    for row in export.to_dict(orient="records"):
        lng = row.pop("lng", None)
        lat = row.pop("lat", None)
        if lng is None or lat is None:
            continue
        for key, value in list(row.items()):
            if pd.isna(value):
                row[key] = None
        features.append(
            {
                "type": "Feature",
                "geometry": {
                    "type": "Point",
                    "coordinates": [float(lng), float(lat)],
                },
                "properties": row,
            }
        )
    return {"type": "FeatureCollection", "features": features}


def _write_geojson(data, output_path):
    output_path.write_text(
        json.dumps(
            _to_features(data),
            ensure_ascii=False,
            indent=2,
            allow_nan=False,
        ),
        encoding="utf-8",
    )


def main():
    root = Path(__file__).resolve().parents[2]
    output_dir = root / "Taipei-City-Dashboard-FE" / "public" / "mapData"
    output_dir.mkdir(parents=True, exist_ok=True)

    taipei = fetch_taipei_onstreet_parking()
    now = pd.Series([taipei["data_time"].max()])
    new_taipei = fetch_new_taipei_onstreet_parking(now.iloc[0])

    _write_geojson(
        taipei,
        output_dir / "tourism_onstreet_parking_taipei.geojson",
    )
    _write_geojson(
        pd.concat([taipei, new_taipei], ignore_index=True),
        output_dir / "tourism_onstreet_parking_metrotaipei.geojson",
    )


if __name__ == "__main__":
    main()
