import json
import math
from pathlib import Path
import re
import sys

import pandas as pd

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))
if str(ROOT / "dags") not in sys.path:
    sys.path.insert(0, str(ROOT / "dags"))

PRICE_RE = re.compile(r"\d+(?:\.\d+)?")
CITY_LABELS = {"taipei": "臺北市", "new_taipei": "新北市"}


def _load_json(path):
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh)


def _load_geojson_features(path):
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh).get("features", [])


def extract_price_value(value):
    numbers = [float(token) for token in PRICE_RE.findall(str(value or ""))]
    return max(numbers) if numbers else None


def parking_lot_type_reference(row):
    checks = [
        ("chargingstation", "充電車位"),
        ("handicap_first", "身障車格"),
        ("pregnancy_first", "孕婦優先車格"),
        ("taxi_onehr_free", "計程車優惠車格"),
        ("child_pickup_area", "家長接送區"),
        ("totalbus", "大客車車位"),
        ("totallargemotor", "大型重機車位"),
        ("totalmotor", "機車停車場"),
        ("totalbike", "自行車停車場"),
    ]
    for column, label in checks:
        if pd.to_numeric(pd.Series([row.get(column)]), errors="coerce").fillna(0).iloc[0] > 0:
            return label
    return "一般停車場"


def finalize_supply_points(data):
    ready = data.copy()
    ready["city_label"] = ready["city"].map(CITY_LABELS)
    ready["district"] = ready["district"].fillna("未知行政區").astype(str).str.strip()
    ready["display_name"] = ready["display_name"].fillna("未命名停車點").astype(str).str.strip()
    ready["address"] = ready["address"].fillna("").astype(str).str.strip()
    ready["facility_kind_label"] = ready["facility_kind"].map(
        {"public_parking": "停車場", "onstreet": "路邊車格"}
    )
    ready["capacity_total"] = pd.to_numeric(ready["capacity_total"], errors="coerce").fillna(0)
    ready["available_total"] = pd.to_numeric(ready["available_total"], errors="coerce")
    ready["available_total"] = ready["available_total"].where(ready["available_total"] >= 0)
    ready["occupied_total"] = pd.to_numeric(ready["occupied_total"], errors="coerce")
    ready["occupied_total"] = ready["occupied_total"].where(
        ready["occupied_total"].notna(),
        (ready["capacity_total"] - ready["available_total"]).clip(lower=0),
    )
    denominator = ready["capacity_total"].replace({0: pd.NA})
    ready["available_rate"] = ready["available_total"] / denominator
    ready["occupancy_rate"] = ready["occupied_total"] / denominator
    ready["price_tier"] = ready["price_value"].apply(
        lambda value: "unknown"
        if pd.isna(value)
        else "free"
        if value <= 0
        else "low"
        if value <= 20
        else "medium"
        if value <= 40
        else "high"
        if value <= 60
        else "premium"
    )
    tier_labels = {
        "unknown": "價格未知",
        "free": "免費/未收費",
        "low": "低價",
        "medium": "中價",
        "high": "高價",
        "premium": "極高價",
    }
    ready["price_tier_label"] = ready["price_tier"].map(tier_labels)
    ready["status_label"] = ready["status_label"].fillna(
        ready["available_total"].map(
            lambda value: "狀態未知"
            if pd.isna(value)
            else "空位"
            if value > 0
            else "已停"
        )
    )
    ready["lng"] = pd.to_numeric(ready["lng"], errors="coerce")
    ready["lat"] = pd.to_numeric(ready["lat"], errors="coerce")
    return ready.dropna(subset=["lng", "lat"])


def convert_twd97_to_wgs84(data, x_col, y_col):
    def convert_point(x, y):
        if pd.isna(x) or pd.isna(y):
            return (None, None)
        a = 6378137.0
        b = 6356752.314245
        lng0 = math.radians(121)
        k0 = 0.9999
        dx = 250000
        dy = 0
        e = (1 - b**2 / a**2) ** 0.5
        x = float(x) - dx
        y = float(y) - dy
        m = y / k0
        mu = m / (
            a
            * (
                1.0
                - e**2 / 4.0
                - 3 * e**4 / 64.0
                - 5 * e**6 / 256.0
            )
        )
        e1 = (1.0 - (1.0 - e**2) ** 0.5) / (1.0 + (1.0 - e**2) ** 0.5)
        j1 = 3 * e1 / 2 - 27 * e1**3 / 32.0
        j2 = 21 * e1**2 / 16 - 55 * e1**4 / 32.0
        j3 = 151 * e1**3 / 96.0
        j4 = 1097 * e1**4 / 512.0
        fp = mu + j1 * math.sin(2 * mu) + j2 * math.sin(4 * mu) + j3 * math.sin(6 * mu) + j4 * math.sin(8 * mu)
        e2 = (e * a / b) ** 2
        c1 = e2 * math.cos(fp) ** 2
        t1 = math.tan(fp) ** 2
        r1 = a * (1 - e**2) / ((1 - e**2 * math.sin(fp) ** 2) ** 1.5)
        n1 = a / (1 - e**2 * math.sin(fp) ** 2) ** 0.5
        d = x / (n1 * k0)
        q1 = n1 * math.tan(fp) / r1
        q2 = d**2 / 2.0
        q3 = (
            5
            + 3 * t1
            + 10 * c1
            - 4 * c1**2
            - 9 * e2
        ) * d**4 / 24.0
        q4 = (
            61
            + 90 * t1
            + 298 * c1
            + 45 * t1**2
            - 3 * c1**2
            - 252 * e2
        ) * d**6 / 720.0
        lat = fp - q1 * (q2 - q3 + q4)
        q5 = d
        q6 = (1 + 2 * t1 + c1) * d**3 / 6.0
        q7 = (
            5
            - 2 * c1
            + 28 * t1
            - 3 * c1**2
            + 8 * e2
            + 24 * t1**2
        ) * d**5 / 120.0
        lng = lng0 + (q5 - q6 + q7) / math.cos(fp)
        return (math.degrees(lng), math.degrees(lat))

    coords = [convert_point(x, y) for x, y in zip(data[x_col], data[y_col])]
    lng = [item[0] for item in coords]
    lat = [item[1] for item in coords]
    return pd.Series(lng), pd.Series(lat)


def _write_geojson_snapshots(gdata, repo_root):
    map_dir = repo_root / "Taipei-City-Dashboard-FE" / "public" / "mapData"
    map_dir.mkdir(parents=True, exist_ok=True)
    snapshots = {
        "parking_supply_points_taipei.geojson": gdata[gdata["city"] == "taipei"],
        "parking_supply_points_metrotaipei.geojson": gdata,
    }
    for file_name, subset in snapshots.items():
        features = []
        for row in subset.to_dict(orient="records"):
            lng = row.pop("lng", None)
            lat = row.pop("lat", None)
            row.pop("wkb_geometry", None)
            if pd.isna(lng) or pd.isna(lat):
                continue
            for key, value in list(row.items()):
                if pd.isna(value):
                    row[key] = None
                elif key == "data_time":
                    row[key] = str(value)
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
        payload = {"type": "FeatureCollection", "features": features}
        (map_dir / file_name).write_text(
            json.dumps(payload, ensure_ascii=False), encoding="utf-8"
        )


def _build_taipei_public(data_dir):
    static_payload = _load_json(data_dir / "D030104-2.json")["data"]
    realtime_payload = _load_json(data_dir / "D030104-1.json")["data"]
    static_rows = static_payload["park"]
    realtime_rows = realtime_payload["park"]
    static = pd.DataFrame(static_rows)
    realtime = pd.DataFrame(realtime_rows)
    static.columns = static.columns.str.lower()
    realtime.columns = realtime.columns.str.lower()
    merged = static.merge(realtime, on="id", how="left", suffixes=("", "_rt"))
    merged["lng"], merged["lat"] = convert_twd97_to_wgs84(merged, "tw97x", "tw97y")
    return pd.DataFrame(
        {
            "supply_id": merged["id"].astype(str).map(lambda value: f"taipei-parking-{value}"),
            "city": "taipei",
            "district": merged["area"],
            "display_name": merged["name"],
            "address": merged["address"],
            "facility_kind": "public_parking",
            "symbol_text": "S",
            "type_reference": merged.apply(
                lambda row: parking_lot_type_reference(
                    {
                        "charging_station": row.get("chargingstation"),
                        "handicap_first_count": row.get("handicap_first"),
                        "pregnancy_first_count": row.get("pregnancy_first"),
                        "taxi_onehr_free_count": row.get("taxi_onehr_free"),
                        "child_pickup_area": row.get("child_pickup_area"),
                        "total_bus": row.get("totalbus"),
                        "total_largemotor": row.get("totallargemotor"),
                        "total_motor": row.get("totalmotor"),
                        "total_bike": row.get("totalbike"),
                    }
                ),
                axis=1,
            ),
            "type_confidence": 1.0,
            "price_raw": merged["payex"],
            "price_value": merged["payex"].map(extract_price_value),
            "capacity_total": pd.to_numeric(merged["totalcar"], errors="coerce"),
            "available_total": pd.to_numeric(merged["availablecar"], errors="coerce").replace(-9, pd.NA),
            "occupied_total": None,
            "status_label": None,
            "data_time": pd.to_datetime(realtime_payload.get("UPDATETIME"), errors="coerce"),
            "lng": merged["lng"],
            "lat": merged["lat"],
        }
    )


def _build_new_taipei_public(data_dir):
    static_payload = _load_json(data_dir / "public_parking_new_tpe.json")
    realtime_payload = _load_json(data_dir / "public_parking_realtime_new_tpe.json")
    static = pd.DataFrame(static_payload)
    realtime = pd.DataFrame(realtime_payload)
    static.columns = static.columns.str.lower()
    realtime.columns = realtime.columns.str.lower()
    merged = static.merge(realtime, on="id", how="left", suffixes=("", "_rt"))
    merged["lng"], merged["lat"] = convert_twd97_to_wgs84(merged, "tw97x", "tw97y")
    return pd.DataFrame(
        {
            "supply_id": merged["id"].astype(str).map(lambda value: f"new_taipei-parking-{value}"),
            "city": "new_taipei",
            "district": merged["area"],
            "display_name": merged["name"],
            "address": merged["address"],
            "facility_kind": "public_parking",
            "symbol_text": "S",
            "type_reference": merged.apply(
                lambda row: parking_lot_type_reference(
                    {
                        "total_bus": row.get("totalbus"),
                        "total_motor": row.get("totalmotor"),
                        "total_bike": row.get("totalbike"),
                    }
                ),
                axis=1,
            ),
            "type_confidence": 1.0,
            "price_raw": merged["payex"],
            "price_value": merged["payex"].map(extract_price_value),
            "capacity_total": pd.to_numeric(merged["totalcar"], errors="coerce"),
            "available_total": pd.to_numeric(merged["availablecar"], errors="coerce").replace(-9, pd.NA),
            "occupied_total": None,
            "status_label": None,
            "data_time": pd.Timestamp.now(tz="Asia/Taipei"),
            "lng": merged["lng"],
            "lat": merged["lat"],
        }
    )


def _build_onstreet(geojson_path, city=None):
    rows = []
    for feature in _load_geojson_features(geojson_path):
        properties = feature.get("properties", {})
        geometry = feature.get("geometry", {})
        coordinates = geometry.get("coordinates", [None, None])
        if city and properties.get("city") != city:
            continue
        rows.append(
            {
                "supply_id": f"{properties.get('city')}-onstreet-{properties.get('parking_id')}",
                "city": properties.get("city"),
                "district": properties.get("district"),
                "display_name": properties.get("name"),
                "address": properties.get("address"),
                "facility_kind": "onstreet",
                "symbol_text": None,
                "type_reference": "路邊停車格",
                "type_confidence": 0.6,
                "price_raw": None,
                "price_value": None,
                "capacity_total": pd.to_numeric(properties.get("total_car"), errors="coerce"),
                "available_total": pd.to_numeric(properties.get("available_car"), errors="coerce"),
                "occupied_total": pd.to_numeric(properties.get("occupied_car"), errors="coerce"),
                "status_label": properties.get("status_label"),
                "data_time": pd.to_datetime(properties.get("data_time"), errors="coerce"),
                "lng": coordinates[0],
                "lat": coordinates[1],
            }
        )
    return pd.DataFrame(rows)


def main():
    data_dir = ROOT / "data"
    fe_map_dir = ROOT.parent / "Taipei-City-Dashboard-FE" / "public" / "mapData"
    ready = pd.concat(
        [
            _build_taipei_public(data_dir),
            _build_new_taipei_public(data_dir),
            _build_onstreet(fe_map_dir / "tourism_onstreet_parking_taipei.geojson"),
            _build_onstreet(fe_map_dir / "tourism_onstreet_parking_metrotaipei.geojson", city="new_taipei"),
        ],
        ignore_index=True,
    )
    snapshots = finalize_supply_points(ready)
    _write_geojson_snapshots(snapshots, ROOT.parent)
    print("Built local parking supply GeoJSON snapshots.")


if __name__ == "__main__":
    main()
