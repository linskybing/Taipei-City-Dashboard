from functools import lru_cache
from io import BytesIO
from pathlib import Path
import hashlib
import re

import pandas as pd
import requests

TAIPEI_EVENTS_URL = (
    "https://data.taipei/api/dataset/6df5ded8-ddfc-413b-93ea-4e26a5a77027/"
    "resource/8deab6c0-005e-4306-89fd-bff881c771a9/download"
)
NEW_TAIPEI_STATS_URL = (
    "https://data.ntpc.gov.tw/api/datasets/"
    "ECEA9D2E-266D-4ADE-B8DD-C88FC0EB9E87/json?page=0&size=20000"
)
OPENDATA_PATH = Path(__file__).resolve().parent / "opendata"
TAIPEI_PARKING_LAW_PREFIXES = ("55", "56")
NEW_TAIPEI_PARKING_LAW_PREFIXES = ("55條", "56條")
TAIPEI_AREA_RE = re.compile(r"([^\d一二三四五六七八九十]+)[一二三四五六七八九十]+區$")


def _tz_now():
    return pd.Timestamp.now(tz="Asia/Taipei")


def _make_hash(prefix, *parts):
    key = "|".join("" if pd.isna(part) else str(part) for part in parts)
    return f"{prefix}_{hashlib.md5(key.encode('utf-8')).hexdigest()[:16]}"


def _categorize_parking(text):
    text = str(text or "")
    if "併排" in text:
        return "併排停車"
    if "身心障礙" in text:
        return "身心障礙停車位違停"
    if "臨時停車" in text:
        return "違規臨時停車"
    if "停車時間位置方式" in text:
        return "停車方式不依規定"
    return "違規停車"


def _normalize_taipei_area_name(area_name):
    text = str(area_name or "").strip()
    matched = TAIPEI_AREA_RE.match(text)
    if matched:
        return f"{matched.group(1)}區"
    return text


@lru_cache(maxsize=4)
def _load_district_code_map(county_name):
    areas = pd.read_csv(
        OPENDATA_PATH / "village" / "VILLAGE_NLSC_1120928.csv",
        usecols=["COUNTYNAME", "TOWNNAME", "TOWNCODE"],
        dtype=str,
    )
    areas = areas[areas["COUNTYNAME"] == county_name][["TOWNCODE", "TOWNNAME"]].drop_duplicates()
    return areas.set_index("TOWNCODE")["TOWNNAME"].to_dict()


def fetch_taipei_illegal_parking_events(data_time=None):
    data_time = data_time or _tz_now()
    raw = requests.get(TAIPEI_EVENTS_URL, timeout=120)
    raw.raise_for_status()
    data = pd.read_csv(BytesIO(raw.content), encoding="cp950", dtype=str).rename(
        columns={
            "西元年": "violation_year",
            "月份": "violation_month",
            "時間": "violation_time",
            "law": "law_code",
            "AreaName": "area_name",
            "地址-行政區域代碼": "district_code",
            "Road": "road",
        }
    )
    data = data[data["law_code"].str.startswith(TAIPEI_PARKING_LAW_PREFIXES, na=False)].copy()
    district_map = _load_district_code_map("臺北市")
    data["city"] = "taipei"
    data["district"] = data["district_code"].map(district_map).fillna(
        data["area_name"].map(_normalize_taipei_area_name)
    )
    data["parking_category"] = data["fact"].map(_categorize_parking)
    data["location_precision"] = "road_only"
    data["data_time"] = data_time
    tokens = zip(
        data["violation_year"],
        data["violation_month"],
        data["violation_time"],
        data["law_code"],
        data["district_code"],
        data["road"],
        range(len(data)),
    )
    data["event_id"] = [_make_hash("tpe_evt", *parts) for parts in tokens]
    return data[
        [
            "event_id",
            "city",
            "district",
            "violation_year",
            "violation_month",
            "violation_time",
            "law_code",
            "fact",
            "parking_category",
            "area_name",
            "district_code",
            "road",
            "location_precision",
            "data_time",
        ]
    ]


def fetch_new_taipei_illegal_parking_stats():
    response = requests.get(NEW_TAIPEI_STATS_URL, timeout=120)
    response.raise_for_status()
    data = pd.DataFrame(response.json()).rename(
        columns={
            "year": "violation_year",
            "month": "violation_month",
            "field1": "violation_type",
            "field2": "law_ref",
            "field3": "vehicle_type",
            "field4": "enforcement_type",
            "quantity": "quantity",
        }
    )
    data = data[
        data["law_ref"].str.startswith(NEW_TAIPEI_PARKING_LAW_PREFIXES, na=False)
    ].copy()
    data["city"] = "new_taipei"
    data["violation_year"] = pd.to_numeric(data["violation_year"], errors="coerce").add(1911).astype("Int64").astype(str)
    data["violation_month"] = data["violation_month"].astype(str).str.zfill(2)
    data["parking_category"] = data["violation_type"].map(_categorize_parking)
    data["quantity"] = pd.to_numeric(data["quantity"], errors="coerce").fillna(0).astype(int)
    data["data_time"] = pd.to_datetime(
        data["violation_year"] + "-" + data["violation_month"] + "-01",
        errors="coerce",
        format="%Y-%m-%d",
    ).dt.tz_localize("Asia/Taipei")
    tokens = zip(
        data["violation_year"],
        data["violation_month"],
        data["violation_type"],
        data["law_ref"],
        data["vehicle_type"],
        data["enforcement_type"],
    )
    data["stat_id"] = [_make_hash("ntpc_stat", *parts) for parts in tokens]
    return data[
        [
            "stat_id",
            "city",
            "violation_year",
            "violation_month",
            "violation_type",
            "law_ref",
            "vehicle_type",
            "enforcement_type",
            "parking_category",
            "quantity",
            "data_time",
        ]
    ]
