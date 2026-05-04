import json
from pathlib import Path

import pandas as pd
import requests

from utils.extract_stage import get_tdx_data
from utils.metrotaipei_parking_supply_common import (
    extract_price_value,
    finalize_supply_points,
    first,
    latest,
    parking_lot_type_reference,
    pick,
    read_table,
    text,
)
from utils.offstreet_parking import NEW_TAIPEI_ONSTREET_URL, TAIPEI_SEGMENT_URL, TAIPEI_STATIC_URL
from utils.transform_geometry import add_point_wkbgeometry_column_to_df


def _log(message):
    print(f"[metrotaipei_parking_supply] {message}", flush=True)


def _build_public_parking(static_data, realtime_data, city_code):
    static_id = pick(set(static_data.columns), "station_id", "id")
    realtime_id = pick(set(realtime_data.columns), "station_id", "id")
    latest_data = latest(
        realtime_data.assign(**{realtime_id: text(realtime_data[realtime_id])}),
        realtime_id,
    )
    merged = static_data.assign(**{static_id: text(static_data[static_id])}).merge(
        latest_data,
        left_on=static_id,
        right_on=realtime_id,
        how="left",
        suffixes=("", "_rt"),
    )
    name_col = pick(set(merged.columns), "name")
    address_col = first(set(merged.columns), "addr", "address") or name_col
    price_raw_col = first(set(merged.columns), "fare_info", "pay_info")
    data_time_col = first(set(merged.columns), "data_time_rt", "data_time")
    type_reference = merged.apply(parking_lot_type_reference, axis=1)
    return pd.DataFrame(
        {
            "supply_id": merged[static_id].map(lambda value: f"{city_code}-parking-{value}"),
            "city": city_code,
            "district": merged[pick(set(merged.columns), "dist", "district")],
            "display_name": merged[name_col],
            "address": merged[address_col],
            "facility_kind": "public_parking",
            "symbol_text": "S",
            "type_reference": type_reference,
            "type_confidence": 1.0,
            "price_raw": merged[price_raw_col] if price_raw_col else None,
            "price_value": [
                extract_price_value(raw)
                for raw in (
                    merged[price_raw_col] if price_raw_col else pd.Series([None] * len(merged))
                )
            ],
            "capacity_total": merged[pick(set(merged.columns), "total_car")],
            "available_total": merged[first(set(merged.columns), "available_car_rt", "available_car")],
            "occupied_total": None,
            "status_label": None,
            "data_time": pd.to_datetime(merged[data_time_col], errors="coerce"),
            "lng": merged[pick(set(merged.columns), "lng", "longitude", "lon")],
            "lat": merged[pick(set(merged.columns), "lat", "latitude")],
        }
    )


def _fetch_taipei_onstreet_meta():
    static_payload = get_tdx_data(TAIPEI_STATIC_URL, output_format="json")
    segment_payload = get_tdx_data(TAIPEI_SEGMENT_URL, output_format="json")
    segment_map = {}
    for segment in segment_payload.get("ParkingSegments", []):
        name_info = segment.get("ParkingSegmentName") or {}
        road_name = name_info.get("Zh_tw") or segment.get("Description") or ""
        segment_map[segment.get("ParkingSegmentID")] = {
            "type_reference": "路邊停車格",
            "type_confidence": 0.4,
            "price_raw": segment.get("FareDescription"),
            "address": road_name,
            "price_value": extract_price_value(segment.get("FareDescription")),
        }
    return pd.DataFrame(
        [
            {"parking_id": str(spot.get("ParkingSpotID")), **segment_map.get(spot.get("ParkingSegmentID"), {})}
            for spot in static_payload.get("ParkingSegmentSpots", [])
        ]
    )


def _fetch_new_taipei_onstreet_meta():
    response = requests.get(NEW_TAIPEI_ONSTREET_URL, timeout=60)
    response.raise_for_status()
    rows = response.json()
    parsed = []
    for row in rows:
        price_raw = " / ".join(
            [value.strip() for value in [str(row.get("pay") or ""), str(row.get("paycash") or "")] if value and value.strip()]
        )
        parsed.append(
            {
                "parking_id": str(row.get("id")),
                "type_reference": str(row.get("name") or "").strip() or "路邊停車格",
                "type_confidence": 1.0 if row.get("name") else 0.5,
                "price_raw": price_raw or None,
                "address_meta": str(row.get("roadname") or "").strip(),
                "price_value": extract_price_value(price_raw),
            }
        )
    return pd.DataFrame(parsed)


def _build_onstreet(current_data, meta_data, city_code):
    parking_id_col = pick(set(current_data.columns), "parking_id", "id")
    merged = current_data.assign(**{parking_id_col: text(current_data[parking_id_col])}).merge(
        meta_data, left_on=parking_id_col, right_on="parking_id", how="left"
    )
    return pd.DataFrame(
        {
            "supply_id": merged[parking_id_col].map(lambda value: f"{city_code}-onstreet-{value}"),
            "city": city_code,
            "district": merged[pick(set(merged.columns), "district")],
            "display_name": merged[pick(set(merged.columns), "name")],
            "address": merged[first(set(merged.columns), "address", "address_meta")] if first(set(merged.columns), "address", "address_meta") else "",
            "facility_kind": "onstreet",
            "symbol_text": None,
            "type_reference": merged["type_reference"].fillna("路邊停車格"),
            "type_confidence": pd.to_numeric(merged["type_confidence"], errors="coerce").fillna(0.5),
            "price_raw": merged["price_raw"] if "price_raw" in merged.columns else None,
            "price_value": pd.to_numeric(merged.get("price_value"), errors="coerce"),
            "capacity_total": merged[pick(set(merged.columns), "total_car")].fillna(1),
            "available_total": merged[first(set(merged.columns), "available_car")],
            "occupied_total": merged[first(set(merged.columns), "occupied_car")],
            "status_label": merged[first(set(merged.columns), "status_label")],
            "data_time": pd.to_datetime(merged[first(set(merged.columns), "data_time")], errors="coerce"),
            "lng": merged[pick(set(merged.columns), "lng", "longitude", "lon")],
            "lat": merged[pick(set(merged.columns), "lat", "latitude")],
        }
    )


def build_metrotaipei_parking_supply(engine):
    _log("Reading current parking tables from PostgreSQL.")
    frames = [
        _build_public_parking(read_table(engine, "tran_parking"), read_table(engine, "tran_parking_capacity_realtime"), "taipei"),
        _build_public_parking(read_table(engine, "tran_parking_new_tpe"), read_table(engine, "tran_parking_capacity_realtime_new_tpe"), "new_taipei"),
        _build_onstreet(read_table(engine, "tran_onstreet_parking_realtime_tpe"), _fetch_taipei_onstreet_meta(), "taipei"),
        _build_onstreet(read_table(engine, "tran_onstreet_parking_realtime_new_tpe"), _fetch_new_taipei_onstreet_meta(), "new_taipei"),
    ]
    ready = finalize_supply_points(pd.concat(frames, ignore_index=True))
    _log(f"Built {len(ready)} metro parking supply points.")
    return ready


def write_geojson_snapshots(gdata, repo_root):
    map_dir = Path(repo_root) / "Taipei-City-Dashboard-FE" / "public" / "mapData"
    map_dir.mkdir(parents=True, exist_ok=True)
    initial_difficulty_district = "萬華區"
    difficulty_initial = gdata[gdata["district"] == initial_difficulty_district]
    difficulty_scopes = {
        "taipei": gdata[gdata["city"] == "taipei"],
        "metrotaipei": gdata,
    }
    snapshots = {
        "parking_supply_points_taipei.geojson": gdata[gdata["city"] == "taipei"],
        "parking_supply_points_metrotaipei.geojson": gdata,
        "parking_public_points_taipei.geojson": gdata[
            (gdata["city"] == "taipei") & (gdata["facility_kind"] == "public_parking")
        ],
        "parking_public_points_metrotaipei.geojson": gdata[
            gdata["facility_kind"] == "public_parking"
        ],
        "parking_onstreet_points_taipei.geojson": gdata[
            (gdata["city"] == "taipei") & (gdata["facility_kind"] == "onstreet")
        ],
        "parking_onstreet_points_metrotaipei.geojson": gdata[
            gdata["facility_kind"] == "onstreet"
        ],
        "parking_difficulty_points_taipei_wanhua.geojson": difficulty_initial[
            difficulty_initial["city"] == "taipei"
        ],
        "parking_difficulty_points_metrotaipei_wanhua.geojson": difficulty_initial,
    }
    manifests = {}
    for scope, scope_data in difficulty_scopes.items():
        chunks = []
        for idx, (district, subset) in enumerate(
            scope_data.groupby("district", dropna=True), start=1
        ):
            file_stem = f"parking_difficulty_points_{scope}_chunk_{idx:02d}"
            snapshots[f"{file_stem}.geojson"] = subset
            chunks.append({"district": district, "index": file_stem})
        manifests[f"parking_difficulty_points_{scope}_manifest.json"] = {
            "initial_district": initial_difficulty_district,
            "chunks": chunks,
        }
    for file_name, subset in snapshots.items():
        geojson_ready = add_point_wkbgeometry_column_to_df(
            subset.copy(), subset["lng"], subset["lat"], from_crs=4326
        ).drop(columns=["wkb_geometry"])
        if "data_time" in geojson_ready.columns:
            geojson_ready["data_time"] = geojson_ready["data_time"].astype(str)
        (map_dir / file_name).write_text(
            geojson_ready.to_json(drop_id=True, ensure_ascii=False), encoding="utf-8"
        )
    for file_name, manifest in manifests.items():
        (map_dir / file_name).write_text(
            json.dumps(manifest, ensure_ascii=False), encoding="utf-8"
        )
