import re

import pandas as pd

from utils.transform_geometry import add_point_wkbgeometry_column_to_df

CITY_LABELS = {"taipei": "臺北市", "new_taipei": "新北市"}
OUTPUT_COLUMNS = [
    "supply_id",
    "city",
    "city_label",
    "district",
    "display_name",
    "address",
    "facility_kind",
    "facility_kind_label",
    "symbol_text",
    "type_reference",
    "type_confidence",
    "price_raw",
    "price_value",
    "price_tier",
    "price_tier_label",
    "capacity_total",
    "available_total",
    "occupied_total",
    "available_rate",
    "occupancy_rate",
    "status_label",
    "data_time",
    "lng",
    "lat",
]
PRICE_RE = re.compile(r"\d+(?:\.\d+)?")
SPECIAL_TYPE_COLUMNS = [
    ("charging_station", "充電車位"),
    ("handicap_first_count", "身障車格"),
    ("pregnancy_first_count", "孕婦優先車格"),
    ("child_pickup_area", "家長接送區"),
    ("taxi_onehr_free_count", "計程車優惠車格"),
]
VEHICLE_TYPE_COLUMNS = [
    ("total_bus", "大客車車位"),
    ("total_largemotor", "大型重機車位"),
    ("total_motor", "機車停車場"),
    ("total_bike", "自行車停車場"),
]


def read_table(engine, table_name):
    return pd.read_sql(f"SELECT * FROM public.{table_name}", engine)


def pick(columns, *candidates):
    for candidate in candidates:
        if candidate in columns:
            return candidate
    raise KeyError(f"Missing required columns {candidates}. Available: {sorted(columns)}")


def first(columns, *candidates):
    for candidate in candidates:
        if candidate in columns:
            return candidate
    return None


def latest(data, key_col):
    time_col = first(set(data.columns), "data_time", "update_time")
    if not time_col:
        return data.drop_duplicates(subset=[key_col], keep="last")
    ranked = data.copy()
    ranked["_sort_time"] = pd.to_datetime(ranked[time_col], errors="coerce")
    ranked = ranked.sort_values("_sort_time", ascending=False)
    return ranked.drop_duplicates(subset=[key_col], keep="first").drop(columns=["_sort_time"])


def numeric(series, fill_value=None):
    values = pd.to_numeric(series, errors="coerce")
    return values.fillna(fill_value) if fill_value is not None else values


def text(series, default=""):
    return series.fillna(default).astype(str).str.strip()


def extract_price_value(*values):
    numbers = []
    for value in values:
        for token in PRICE_RE.findall(str(value or "")):
            numbers.append(float(token))
    return max(numbers) if numbers else None


def price_tier(price_value):
    if pd.isna(price_value):
        return "unknown", "價格未知"
    if price_value <= 0:
        return "free", "免費/未收費"
    if price_value <= 20:
        return "low", "低價"
    if price_value <= 40:
        return "medium", "中價"
    if price_value <= 60:
        return "high", "高價"
    return "premium", "極高價"


def normalize_type(label):
    raw = str(label or "").replace("臺", "台")
    if not raw:
        return "未標示"
    checks = [
        ("家長接送", "家長接送"),
        ("身障", "身障"),
        ("身心障礙", "身障"),
        ("裝卸", "裝卸貨"),
        ("卸貨", "裝卸貨"),
        ("大型車", "大型車"),
        ("大客車", "大型車"),
        ("機車", "機車"),
        ("彈性", "彈性共用"),
        ("共用", "彈性共用"),
        ("計程車", "計程車"),
        ("限時", "限時/時段"),
        ("時段", "限時/時段"),
        ("禁停", "限時/時段"),
        ("警用", "公用/警用"),
        ("垃圾車", "公用/警用"),
        ("公務", "公用/警用"),
        ("汽車", "汽車"),
        ("小型車", "汽車"),
    ]
    for needle, normalized in checks:
        if needle in raw:
            return normalized
    return raw


def parking_lot_type_reference(row):
    for column, label in SPECIAL_TYPE_COLUMNS:
        if float(row.get(column) or 0) > 0:
            return label
    for column, label in VEHICLE_TYPE_COLUMNS:
        if float(row.get(column) or 0) > 0:
            return label
    return "一般停車場"


def status_label(available_total, capacity_total, raw_status=None):
    if raw_status:
        return str(raw_status).strip()
    if pd.isna(available_total):
        return "狀態未知"
    if available_total > 0:
        return "空位"
    if capacity_total > 0:
        return "已停"
    return "不可用"


def finalize_supply_points(data):
    data["city_label"] = data["city"].map(CITY_LABELS)
    data["district"] = text(data["district"], "未知行政區")
    data["display_name"] = text(data["display_name"], "未命名停車點")
    data["address"] = text(data["address"])
    data["data_time"] = pd.to_datetime(data["data_time"], errors="coerce")
    data["facility_kind_label"] = data["facility_kind"].map(
        {"public_parking": "停車場", "onstreet": "路邊車格"}
    )
    data["symbol_text"] = data["symbol_text"].where(data["symbol_text"].notna(), None)
    data["capacity_total"] = numeric(data["capacity_total"], 0)
    data["available_total"] = numeric(data["available_total"])
    data.loc[data["available_total"] < 0, "available_total"] = pd.NA
    data["occupied_total"] = numeric(data["occupied_total"])
    data["occupied_total"] = data["occupied_total"].where(
        data["occupied_total"].notna(),
        (data["capacity_total"] - data["available_total"]).clip(lower=0),
    )
    denominator = data["capacity_total"].replace({0: pd.NA})
    data["available_rate"] = data["available_total"] / denominator
    data["occupancy_rate"] = data["occupied_total"] / denominator
    tiers = data["price_value"].apply(price_tier)
    data["price_tier"] = tiers.map(lambda value: value[0])
    data["price_tier_label"] = tiers.map(lambda value: value[1])
    data["status_label"] = [
        status_label(available, capacity, raw_status)
        for available, capacity, raw_status in zip(
            data["available_total"], data["capacity_total"], data["status_label"]
        )
    ]
    data["lng"] = numeric(data["lng"])
    data["lat"] = numeric(data["lat"])
    data = data.dropna(subset=["lng", "lat"])
    gdata = add_point_wkbgeometry_column_to_df(data, data["lng"], data["lat"], from_crs=4326)
    return gdata[OUTPUT_COLUMNS + ["wkb_geometry"]]
