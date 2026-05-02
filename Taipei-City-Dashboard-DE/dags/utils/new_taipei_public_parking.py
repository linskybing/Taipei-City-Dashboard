import json

import pandas as pd
from numpy import nan

STATIC_URL = (
    "https://data.ntpc.gov.tw/api/datasets/"
    "B1464EF0-9C7C-4A6F-ABF7-6BDF32847E68/json?page=0&size=5000"
)
REALTIME_URL = (
    "https://data.ntpc.gov.tw/api/datasets/"
    "e09b35a5-a738-48cc-b0f5-570b67ad9c78/json?page=0&size=5000"
)
TYPE_MAP = {"1": "動態回傳剩餘車位數", "2": "靜態"}
STATIC_COLUMNS = [
    "data_time",
    "station_id",
    "dist",
    "name",
    "data_return_type",
    "owner_type",
    "summary",
    "addr",
    "tel",
    "pay_info",
    "opening_time",
    "total_car",
    "total_motor",
    "total_bike",
    "total_bus",
    "total_largemotor",
    "pregnancy_first_count",
    "handicap_first_count",
    "taxi_onehr_free_count",
    "aed_equipment",
    "cellsignal_enhancement",
    "accessibility_elevator",
    "phone_charge",
    "child_pickup_area",
    "charging_station",
    "fare_info",
    "entrance_coord",
    "x_97",
    "y_97",
]
REALTIME_COLUMNS = [
    "data_time",
    "station_id",
    "available_car",
    "available_motor",
    "available_bus",
    "charge_spot_count",
    "standby_spot_count",
]


def _log(message):
    print(f"[new_taipei_public_parking] {message}", flush=True)


def _text_column(data, column):
    values = data[column] if column in data.columns else pd.Series("", index=data.index)
    return values.fillna("").astype(str).str.strip()


def _numeric_column(data, column):
    values = data[column] if column in data.columns else pd.Series(nan, index=data.index)
    return pd.to_numeric(values, errors="coerce")


def _get_data_time():
    from utils.get_time import get_tpe_now_time_str
    from utils.transform_time import convert_str_to_time_format

    return convert_str_to_time_format(
        pd.Series([get_tpe_now_time_str(is_with_tz=True)])
    ).iloc[0]


def _load_rows(file_name, url):
    from utils.extract_stage import download_file

    _log(f"Downloading {file_name} from {url}")
    with open(download_file(file_name, url, is_proxy=False), encoding="utf-8-sig") as file:
        payload = json.load(file)
    if isinstance(payload, dict):
        rows = payload.get("value", [])
    else:
        rows = payload
    _log(f"Loaded {len(rows)} rows from {file_name}")
    return rows


def fetch_new_taipei_public_parking_static():
    data = pd.DataFrame(_load_rows("public_parking_new_tpe.json", STATIC_URL))
    if data.empty:
        raise ValueError("New Taipei static public parking API returned no rows.")

    data.columns = data.columns.str.lower()
    data = data.rename(
        columns={
            "id": "station_id",
            "area": "dist",
            "name": "name",
            "type": "data_return_type",
            "summary": "summary",
            "address": "addr",
            "tel": "tel",
            "payex": "pay_info",
            "servicetime": "opening_time",
            "tw97x": "x_97",
            "tw97y": "y_97",
            "totalcar": "total_car",
            "totalmotor": "total_motor",
            "totalbike": "total_bike",
        }
    )
    data_time = _get_data_time()
    for col in ["station_id", "dist", "name", "summary", "addr", "tel", "pay_info", "opening_time"]:
        data[col] = _text_column(data, col)
    for col in ["x_97", "y_97"]:
        data[col] = _numeric_column(data, col)
    for col in ["total_car", "total_motor", "total_bike"]:
        data[col] = _numeric_column(data, col).fillna(0)
    data["data_time"] = data_time
    data["data_return_type"] = (
        _text_column(data, "data_return_type").map(TYPE_MAP).fillna("未提供")
    )
    data["owner_type"] = ""
    data["total_bus"] = 0
    data["total_largemotor"] = 0
    data["pregnancy_first_count"] = 0
    data["handicap_first_count"] = 0
    data["taxi_onehr_free_count"] = 0
    data["aed_equipment"] = 0
    data["cellsignal_enhancement"] = 0
    data["accessibility_elevator"] = 0
    data["phone_charge"] = 0
    data["child_pickup_area"] = 0
    data["charging_station"] = 0
    data["fare_info"] = data["pay_info"]
    data["entrance_coord"] = ""
    _log(f"Prepared {len(data)} static parking rows.")
    return data[STATIC_COLUMNS]


def fetch_new_taipei_public_parking_realtime():
    data = pd.DataFrame(_load_rows("public_parking_realtime_new_tpe.json", REALTIME_URL))
    if data.empty:
        raise ValueError("New Taipei realtime public parking API returned no rows.")

    data.columns = data.columns.str.lower()
    data = data.rename(columns={"id": "station_id", "availablecar": "available_car"})
    data["station_id"] = data["station_id"].fillna("").astype(str)
    data["available_car"] = pd.to_numeric(data["available_car"], errors="coerce")
    data.loc[data["available_car"] == -9, "available_car"] = nan
    data["available_motor"] = nan
    data["available_bus"] = nan
    data["charge_spot_count"] = 0
    data["standby_spot_count"] = 0
    data["data_time"] = _get_data_time()
    _log(f"Prepared {len(data)} realtime parking rows.")
    return data[REALTIME_COLUMNS]
