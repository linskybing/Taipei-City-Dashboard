import json
from pathlib import Path
import sys

import pandas as pd

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from dags.utils.illegal_parking_devices import (
    fetch_metrotaipei_illegal_parking_devices,
    fetch_taipei_illegal_parking_devices,
)


EXPORT_COLUMNS = [
    "device_id",
    "city",
    "district",
    "location_name",
    "item",
    "device_type",
    "location_precision",
    "data_time",
]
CITY_LABELS = {"taipei": "臺北市", "new_taipei": "新北市"}


def _to_geojson(data):
    export = data[EXPORT_COLUMNS + ["lng", "lat"]].copy()
    export["data_time"] = export["data_time"].astype(str)
    export["city_label"] = export["city"].map(CITY_LABELS).fillna(export["city"])
    features = []
    for row in export.to_dict(orient="records"):
        lng = row.pop("lng", None)
        lat = row.pop("lat", None)
        if pd.isna(lng) or pd.isna(lat):
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
        json.dumps(_to_geojson(data), ensure_ascii=False, indent=2, allow_nan=False),
        encoding="utf-8",
    )


def main():
    output_dir = ROOT.parent / "Taipei-City-Dashboard-FE" / "public" / "mapData"
    output_dir.mkdir(parents=True, exist_ok=True)
    data_time = pd.Timestamp.now(tz="Asia/Taipei")
    taipei = fetch_taipei_illegal_parking_devices(data_time)
    metrotaipei = fetch_metrotaipei_illegal_parking_devices(data_time)
    _write_geojson(taipei, output_dir / "illegal_parking_devices_taipei.geojson")
    _write_geojson(
        metrotaipei,
        output_dir / "illegal_parking_devices_metrotaipei.geojson",
    )


if __name__ == "__main__":
    main()
