from functools import lru_cache
from pathlib import Path
import re

import pandas as pd
import requests

TAIPEI_STATIC_URL = (
    "https://tdx.transportdata.tw/api/basic/v1/"
    "Parking/OnStreet/ParkingSpot/City/Taipei?%24format=JSON"
)
TAIPEI_SEGMENT_URL = (
    "https://tdx.transportdata.tw/api/basic/v1/"
    "Parking/OnStreet/ParkingSegment/City/Taipei?%24format=JSON"
)
TAIPEI_AVAILABILITY_URL = (
    "https://tdx.transportdata.tw/api/basic/v1/"
    "Parking/OnStreet/ParkingSpotAvailability/City/Taipei?%24format=JSON"
)
NEW_TAIPEI_ONSTREET_URL = (
    "https://data.ntpc.gov.tw/api/datasets/"
    "54a507c4-c038-41b5-bf60-bbecb9d052c6/json?page=0&size=50000"
)
OPENDATA_PATH = Path(__file__).resolve().parent / "opendata"

TAIPEI_STATUS_MAP = {0: ("狀態未知", -1), 1: ("空位", 1), 2: ("已停", 0), 3: ("不可用", -1)}
NEW_TAIPEI_STATUS_MAP = {
    "1": ("空位", 1),
    "2": ("已停", 0),
    "3": ("狀態未知", -1),
    "4": ("不可用", -1),
    "5": ("不可用", -1),
}
TAIPEI_SEGMENT_SUFFIXES = ("臨時平面停車場", "大車場", "北側", "南側", "東側", "西側")
TAIPEI_TRAILING_ZONE_RE = re.compile(r"[A-Z]$")
TAIPEI_NAMELESS_LANE_RE = re.compile(r"\d+號旁無名巷$")
TAIPEI_UNIT_NUMBER_RE = re.compile(r"([一二三四五六七八九十]+)(?=[段巷弄])")


def _get_json(url, timeout=60):
    response = requests.get(url, timeout=timeout)
    response.raise_for_status()
    return response.json()


def _parse_time_series(values):
    parsed = pd.to_datetime(values, errors="coerce")
    if getattr(parsed.dt, "tz", None) is None:
        return parsed.dt.tz_localize("Asia/Taipei")
    return parsed.dt.tz_convert("Asia/Taipei")


def _zh_number_to_int(token):
    digits = {"一": 1, "二": 2, "三": 3, "四": 4, "五": 5, "六": 6, "七": 7, "八": 8, "九": 9}
    if token == "十":
        return 10
    if "十" not in token:
        return int("".join(str(digits[ch]) for ch in token))
    left, _, right = token.partition("十")
    tens = digits[left] if left else 1
    ones = digits[right] if right else 0
    return tens * 10 + ones


def _normalize_taipei_road_text(text):
    if not text:
        return ""
    text = str(text).strip().replace("臺", "台")
    text = TAIPEI_UNIT_NUMBER_RE.sub(lambda m: str(_zh_number_to_int(m.group(1))), text)
    text = TAIPEI_TRAILING_ZONE_RE.sub("", text)
    text = TAIPEI_NAMELESS_LANE_RE.sub("", text)
    for suffix in TAIPEI_SEGMENT_SUFFIXES:
        if text.endswith(suffix):
            text = text[: -len(suffix)]
    return text.strip()


@lru_cache(maxsize=1)
def _load_taipei_road_map():
    roads = pd.read_csv(OPENDATA_PATH / "街道" / "opendata109road.csv", dtype=str)
    roads.columns = ["city", "site_id", "road"]
    roads = roads[roads["city"].isin(["臺北市", "台北市"])].copy()
    roads["district"] = roads["site_id"].str.replace(r"^(臺北市|台北市)", "", regex=True)
    roads["road"] = roads["road"].str.strip()
    unique_roads = roads.groupby("road")["district"].nunique()
    unique_roads = unique_roads[unique_roads == 1].index
    roads = roads[roads["road"].isin(unique_roads)].drop_duplicates("road")
    return {
        _normalize_taipei_road_text(road): district
        for road, district in roads.set_index("road")["district"].to_dict().items()
    }


def _resolve_taipei_district(road_map, *candidates):
    road_keys = sorted(road_map.keys(), key=len, reverse=True)
    for candidate in candidates:
        text = _normalize_taipei_road_text(candidate)
        if not text:
            continue
        if text in road_map:
            return road_map[text]
        for key in road_keys:
            if text.startswith(key):
                return road_map[key]
    return None


@lru_cache(maxsize=1)
def _load_taipei_district_geometries():
    from shapely import wkt
    from shapely.ops import unary_union

    areas = pd.read_csv(
        OPENDATA_PATH / "village" / "VILLAGE_NLSC_1120928.csv",
        usecols=["COUNTYNAME", "TOWNNAME", "geometry"],
        dtype=str,
    )
    areas = areas[areas["COUNTYNAME"].isin(["臺北市", "台北市"])]
    return {
        district: unary_union(group["geometry"].dropna().map(wkt.loads).tolist())
        for district, group in areas.groupby("TOWNNAME")
    }


def _fill_taipei_district_by_point(data):
    from shapely.geometry import Point

    district_geometries = _load_taipei_district_geometries()
    missing = data["district"].isna()
    if not missing.any():
        return data
    for idx, row in data.loc[missing, ["lng", "lat"]].dropna().iterrows():
        point = Point(row["lng"], row["lat"])
        for district, geometry in district_geometries.items():
            if geometry.covers(point):
                data.at[idx, "district"] = district
                break
    return data


@lru_cache(maxsize=1)
def _load_new_taipei_area_map():
    areas = pd.read_csv(
        OPENDATA_PATH / "village" / "VILLAGE_NLSC_1120928.csv",
        usecols=["COUNTYNAME", "TOWNNAME", "TOWNCODE"],
        dtype=str,
    )
    areas = areas[areas["COUNTYNAME"] == "新北市"][["TOWNCODE", "TOWNNAME"]].drop_duplicates()
    return areas.set_index("TOWNCODE")["TOWNNAME"].to_dict()


def _build_name(road_name, parking_id):
    road = str(road_name).strip() or "未知路段"
    return f"{road} 車格 {parking_id}"


def _finalize_onstreet_data(data, city, data_time):
    data["parking_id"] = data["parking_id"].astype(str)
    data["city"] = city
    data["district"] = data["district"].fillna("未知行政區").astype(str).str.strip()
    data["name"] = data["name"].astype(str).str.strip()
    data["address"] = data["address"].fillna("").astype(str).str.strip()
    data["total_car"] = 1
    data["available_car_raw"] = pd.to_numeric(data["available_car_raw"], errors="coerce")
    data["available_car"] = data["available_car_raw"].where(data["available_car_raw"] >= 0)
    data["occupied_car"] = (1 - data["available_car"]).where(data["available_car"].notna())
    data["occupancy_rate"] = data["occupied_car"]
    data["data_time"] = data_time
    data["lng"] = pd.to_numeric(data["lng"], errors="coerce")
    data["lat"] = pd.to_numeric(data["lat"], errors="coerce")
    data = data.dropna(subset=["lng", "lat"])
    return data[
        [
            "parking_id",
            "city",
            "district",
            "name",
            "address",
            "total_car",
            "available_car_raw",
            "available_car",
            "occupied_car",
            "occupancy_rate",
            "status_label",
            "data_time",
            "lng",
            "lat",
        ]
    ]


def fetch_taipei_onstreet_parking():
    from utils.extract_stage import get_tdx_data

    road_map = _load_taipei_road_map()
    static_payload = get_tdx_data(TAIPEI_STATIC_URL, output_format="json")
    segment_payload = get_tdx_data(TAIPEI_SEGMENT_URL, output_format="json")
    availability_payload = get_tdx_data(TAIPEI_AVAILABILITY_URL, output_format="json")

    segment_map = {}
    for segment in segment_payload.get("ParkingSegments", []):
        segment_name = (
            (segment.get("ParkingSegmentName") or {}).get("Zh_tw")
            or segment.get("Description")
            or ""
        )
        section = segment.get("RoadSection") or {}
        segment_map[segment.get("ParkingSegmentID")] = {
            "name": segment_name,
            "start": section.get("Start"),
            "end": section.get("End"),
        }

    static_rows = []
    for spot in static_payload.get("ParkingSegmentSpots", []):
        parking_id = spot.get("ParkingSpotID")
        position = spot.get("Position") or {}
        segment_info = segment_map.get(spot.get("ParkingSegmentID"), {})
        road_name = segment_info.get("name", "")
        static_rows.append(
            {
                "parking_id": parking_id,
                "district": _resolve_taipei_district(
                    road_map,
                    road_name,
                    segment_info.get("start"),
                    segment_info.get("end"),
                ),
                "name": _build_name(road_name, parking_id),
                "address": road_name,
                "lng": position.get("PositionLon"),
                "lat": position.get("PositionLat"),
            }
        )

    availability_rows = []
    for spot in availability_payload.get("CurbSpotParkingAvailabilities", []):
        status_label, available_raw = TAIPEI_STATUS_MAP.get(
            spot.get("SpotStatus"), ("狀態未知", -1)
        )
        availability_rows.append(
            {
                "parking_id": spot.get("ParkingSpotID"),
                "available_car_raw": available_raw,
                "status_label": status_label,
                "collect_time": spot.get("DataCollectTime"),
            }
        )

    availability = pd.DataFrame(availability_rows)
    data_time = pd.Timestamp.now(tz="Asia/Taipei")
    if availability_payload.get("UpdateTime"):
        data_time = _parse_time_series(pd.Series([availability_payload["UpdateTime"]])).iloc[0]
    elif not availability.empty and availability["collect_time"].notna().any():
        data_time = _parse_time_series(availability["collect_time"]).dropna().max()

    merged = pd.DataFrame(static_rows).merge(
        availability[["parking_id", "available_car_raw", "status_label"]],
        on="parking_id",
        how="left",
    )
    merged = _fill_taipei_district_by_point(merged)
    merged["available_car_raw"] = merged["available_car_raw"].fillna(-1)
    merged["status_label"] = merged["status_label"].fillna("狀態未知")
    return _finalize_onstreet_data(merged, "taipei", data_time)


def fetch_new_taipei_onstreet_parking(data_time):
    area_map = _load_new_taipei_area_map()
    rows = _get_json(NEW_TAIPEI_ONSTREET_URL)
    parsed_rows = []
    for row in rows:
        status_label, available_raw = NEW_TAIPEI_STATUS_MAP.get(
            str(row.get("parkingstatus")), ("狀態未知", -1)
        )
        parking_id = row.get("id")
        road_name = row.get("roadname", "")
        parsed_rows.append(
            {
                "parking_id": parking_id,
                "district": area_map.get(str(row.get("areacode"))),
                "name": _build_name(road_name, parking_id),
                "address": road_name,
                "available_car_raw": available_raw,
                "status_label": status_label,
                "lng": row.get("longitude"),
                "lat": row.get("latitude"),
            }
        )
    return _finalize_onstreet_data(pd.DataFrame(parsed_rows), "new_taipei", data_time)
