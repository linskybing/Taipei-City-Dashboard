from functools import lru_cache
from io import StringIO
from pathlib import Path
import hashlib
import re

import pandas as pd
import requests

NEW_TAIPEI_DEVICE_URL = (
    "https://data.ntpc.gov.tw/api/datasets/"
    "BB59A616-3572-4E92-9D09-01BF422057A6/json?page=0&size=1000"
)
TAIPEI_DEVICE_URL = "https://td.police.gov.taipei/cp.aspx?n=6FEDE1F9DBFD656E"
OPENDATA_PATH = Path(__file__).resolve().parent / "opendata"
DEVICE_HEADERS = {"編號", "設置位置", "取締項目", "類型"}


def _tz_now():
    return pd.Timestamp.now(tz="Asia/Taipei")


def _make_hash(prefix, *parts):
    key = "|".join("" if pd.isna(part) else str(part) for part in parts)
    return f"{prefix}_{hashlib.md5(key.encode('utf-8')).hexdigest()[:16]}"


def _primary_district(text):
    return str(text or "").split("、")[0].strip()


@lru_cache(maxsize=2)
def _load_district_centroids(county_name):
    areas = pd.read_csv(
        OPENDATA_PATH / "village" / "VILLAGE_NLSC_1120928.csv",
        usecols=["COUNTYNAME", "TOWNNAME", "geometry"],
        dtype=str,
    )
    areas = areas[areas["COUNTYNAME"] == county_name]
    centroids = {}
    try:
        from shapely import wkt
        from shapely.ops import unary_union

        for district, group in areas.groupby("TOWNNAME"):
            geometry = unary_union(group["geometry"].dropna().map(wkt.loads).tolist()).centroid
            centroids[district] = (geometry.x, geometry.y)
        return centroids
    except Exception:
        pass
    coord_re = re.compile(r"(-?\d+(?:\.\d+)?)\s+(-?\d+(?:\.\d+)?)")
    for district, group in areas.groupby("TOWNNAME"):
        xs = []
        ys = []
        for geometry in group["geometry"].dropna():
            for x_val, y_val in coord_re.findall(geometry):
                xs.append(float(x_val))
                ys.append(float(y_val))
        if xs and ys:
            centroids[district] = (sum(xs) / len(xs), sum(ys) / len(ys))
    return centroids


def _fill_taipei_geocode_or_centroid(data):
    try:
        from utils.transform_address import get_addr_xy_parallel

        lng, lat = get_addr_xy_parallel(data["query_address"].tolist(), sleep_time=0)
    except Exception:
        lng = [None] * len(data)
        lat = [None] * len(data)
    data["lng"] = pd.to_numeric(lng, errors="coerce")
    data["lat"] = pd.to_numeric(lat, errors="coerce")
    data["location_precision"] = pd.Series(["geocoded_query"] * len(data))
    centroids = _load_district_centroids("臺北市")
    missing = data["lng"].isna() | data["lat"].isna()
    for idx, row in data.loc[missing, ["district"]].iterrows():
        district = _primary_district(row["district"])
        if district in centroids:
            data.at[idx, "lng"] = centroids[district][0]
            data.at[idx, "lat"] = centroids[district][1]
            data.at[idx, "location_precision"] = "district_centroid"
    return data


def _strip_html_tags(text):
    return re.sub(r"<.*?>", "", str(text or "")).replace("&nbsp;", " ").strip()


def _parse_taipei_device_rows(html):
    pattern = re.compile(
        r"<caption><strong>(?P<district>.*?)</strong></caption><tbody>(?P<tbody>.*?)</tbody>",
        re.S,
    )
    row_pattern = re.compile(r"<tr>(.*?)</tr>", re.S)
    cell_pattern = re.compile(r"<td[^>]*>(.*?)</td>", re.S)
    rows = []
    for match in pattern.finditer(html):
        district = _strip_html_tags(match.group("district"))
        for row_html in row_pattern.findall(match.group("tbody"))[1:]:
            cells = [_strip_html_tags(cell) for cell in cell_pattern.findall(row_html)]
            if len(cells) != 4:
                continue
            rows.append(
                {
                    "district": district,
                    "編號": cells[0],
                    "設置位置": cells[1],
                    "取締項目": cells[2],
                    "類型": cells[3],
                }
            )
    return rows


def fetch_taipei_illegal_parking_devices(data_time=None):
    data_time = data_time or _tz_now()
    response = requests.get(TAIPEI_DEVICE_URL, timeout=120)
    response.raise_for_status()
    html = response.text
    rows = []
    for row in _parse_taipei_device_rows(html):
        item = str(row.get("取締項目", ""))
        device_type = str(row.get("類型", ""))
        if "違規停車" not in item and "違規停車" not in device_type:
            continue
        district = str(row.get("district", "")).strip()
        location = str(row.get("設置位置", "")).strip()
        rows.append(
            {
                "city": "taipei",
                "district": district,
                "location_name": location,
                "item": item,
                "device_type": device_type,
                "query_address": f"臺北市{_primary_district(district)}{location}",
                "data_time": data_time,
            }
        )
    data = _fill_taipei_geocode_or_centroid(pd.DataFrame(rows))
    data["device_id"] = [
        _make_hash("tpe_dev", district, location_name, item)
        for district, location_name, item in zip(data["district"], data["location_name"], data["item"])
    ]
    return data[
        [
            "device_id",
            "city",
            "district",
            "location_name",
            "item",
            "device_type",
            "location_precision",
            "data_time",
            "lng",
            "lat",
        ]
    ]


def fetch_new_taipei_illegal_parking_devices(data_time=None):
    data_time = data_time or _tz_now()
    response = requests.get(NEW_TAIPEI_DEVICE_URL, timeout=120)
    response.raise_for_status()
    data = pd.DataFrame(response.json()).rename(
        columns={"seqno": "source_seq", "location": "location_name", "latitude": "lat", "longitude": "lng"}
    )
    data["city"] = "new_taipei"
    data["district"] = data["location_name"].astype(str).str.extract(r"^(.+?區)")
    data["device_type"] = "違規停車自動偵測"
    data["location_precision"] = "source_point"
    data["data_time"] = data_time
    data["device_id"] = [
        _make_hash("ntpc_dev", seq, location_name, item)
        for seq, location_name, item in zip(data["source_seq"], data["location_name"], data["item"])
    ]
    return data[
        [
            "device_id",
            "city",
            "district",
            "location_name",
            "item",
            "device_type",
            "location_precision",
            "data_time",
            "lng",
            "lat",
        ]
    ]


def fetch_metrotaipei_illegal_parking_devices(data_time=None):
    data_time = data_time or _tz_now()
    combined = pd.concat(
        [
            fetch_taipei_illegal_parking_devices(data_time),
            fetch_new_taipei_illegal_parking_devices(data_time),
        ],
        ignore_index=True,
    )
    return combined.dropna(subset=["lng", "lat"]).copy()
