import gzip
import xml.etree.ElementTree as ET

import geopandas as gpd
import pandas as pd
import requests
from shapely import wkt
from shapely.geometry import LineString, Point
from shapely.ops import substring

DISTANCE_CRS = 3826


def classify_congestion(speed):
    if pd.isna(speed):
        return 0, "資料不足"
    if float(speed) < 15:
        return 3, "壅塞"
    if float(speed) < 25:
        return 2, "車多"
    return 1, "順暢"


def classify_vd_congestion(speed, total_vol):
    if pd.isna(speed):
        return 0, "資料不足"
    if float(speed) <= 0 and (pd.isna(total_vol) or float(total_vol) <= 0):
        return 0, "資料不足"
    return classify_congestion(speed)


def fetch_taipei_vd_sections(url, timeout=60):
    response = requests.get(url, timeout=timeout)
    response.raise_for_status()
    root = ET.fromstring(gzip.decompress(response.content))
    ns = {"vd": "http://www.iii.org.tw/dax/vd"}
    exchange_time = root.findtext("vd:ExchangeTime", default=None, namespaces=ns)
    rows = []
    for section in root.findall(".//vd:SectionData", ns):
        start_x = pd.to_numeric(section.findtext("vd:StartWgsX", namespaces=ns), errors="coerce")
        start_y = pd.to_numeric(section.findtext("vd:StartWgsY", namespaces=ns), errors="coerce")
        end_x = pd.to_numeric(section.findtext("vd:EndWgsX", namespaces=ns), errors="coerce")
        end_y = pd.to_numeric(section.findtext("vd:EndWgsY", namespaces=ns), errors="coerce")
        if pd.isna(start_x) or pd.isna(start_y) or pd.isna(end_x) or pd.isna(end_y):
            continue
        avg_speed = pd.to_numeric(section.findtext("vd:AvgSpd", namespaces=ns), errors="coerce")
        total_vol = pd.to_numeric(section.findtext("vd:TotalVol", namespaces=ns), errors="coerce")
        status_code, status_label = classify_vd_congestion(avg_speed, total_vol)
        rows.append(
            {
                "city": "taipei",
                "city_label": "臺北市",
                "segment_id": section.findtext("vd:SectionId", default="", namespaces=ns),
                "section_id": section.findtext("vd:SectionId", default="", namespaces=ns),
                "section_name": section.findtext("vd:SectionName", default="", namespaces=ns),
                "route_name": None,
                "source_type": "VD",
                "avg_speed": avg_speed,
                "avg_occ": pd.to_numeric(section.findtext("vd:AvgOcc", namespaces=ns), errors="coerce"),
                "total_vol": total_vol,
                "sample_count": total_vol,
                "moe_level": pd.to_numeric(section.findtext("vd:MOELevel", namespaces=ns), errors="coerce"),
                "status_code": status_code,
                "status_label": status_label,
                "segment_length_m": None,
                "data_time": exchange_time,
                "geometry": LineString([(float(start_x), float(start_y)), (float(end_x), float(end_y))]),
            }
        )
    return gpd.GeoDataFrame(rows, geometry="geometry", crs="EPSG:4326")


def build_route_shape_lookup(shape_rows, stop_rows):
    shape_records = []
    for row in shape_rows:
        geometry_wkt = row.get("Geometry")
        if not geometry_wkt:
            continue
        shape_records.append(
            {
                "routeuid": row.get("RouteUID"),
                "direction": int(row.get("Direction", 0)),
                "route_name": (row.get("RouteName") or {}).get("Zh_tw") or row.get("RouteID"),
                "geometry": wkt.loads(geometry_wkt),
            }
        )
    shape_gdf = gpd.GeoDataFrame(shape_records, geometry="geometry", crs="EPSG:4326")
    shape_gdf = shape_gdf.drop_duplicates(subset=["routeuid", "direction"]).to_crs(epsg=DISTANCE_CRS)

    stop_records = []
    for row in stop_rows:
        for stop in row.get("Stops", []):
            position = stop.get("StopPosition") or {}
            stop_records.append(
                {
                    "routeuid": row.get("RouteUID"),
                    "direction": int(row.get("Direction", 0)),
                    "geometry": Point(position.get("PositionLon"), position.get("PositionLat")),
                }
            )
    if stop_records:
        stop_gdf = gpd.GeoDataFrame(stop_records, geometry="geometry", crs="EPSG:4326").to_crs(epsg=DISTANCE_CRS)
        stop_groups = (
            stop_gdf.groupby(["routeuid", "direction"])["geometry"]
            .apply(list)
            .to_dict()
        )
    else:
        stop_groups = {}

    route_lookup = {}
    for row in shape_gdf.itertuples(index=False):
        route_lookup[(row.routeuid, int(row.direction))] = {
            "route_name": row.route_name,
            "line": row.geometry,
            "stops": stop_groups.get((row.routeuid, int(row.direction)), []),
        }
    return route_lookup


def estimate_new_taipei_segments(
    realtime_rows,
    shape_rows,
    stop_rows,
    segment_length_m=120,
    stop_buffer_m=20,
    max_match_distance_m=60,
):
    route_lookup = build_route_shape_lookup(shape_rows, stop_rows)
    aggregates = {}
    for row in realtime_rows:
        position = row.get("BusPosition") or {}
        route_key = (row.get("RouteUID"), int(row.get("Direction", 0)))
        route = route_lookup.get(route_key)
        if not route:
            continue
        speed = pd.to_numeric(row.get("Speed"), errors="coerce")
        lon = pd.to_numeric(position.get("PositionLon"), errors="coerce")
        lat = pd.to_numeric(position.get("PositionLat"), errors="coerce")
        if pd.isna(speed) or pd.isna(lon) or pd.isna(lat):
            continue
        bus_point = gpd.GeoSeries([Point(float(lon), float(lat))], crs="EPSG:4326").to_crs(epsg=DISTANCE_CRS).iloc[0]
        if float(speed) < 3 and any(bus_point.distance(stop) <= stop_buffer_m for stop in route["stops"]):
            continue
        snapped_distance = route["line"].project(bus_point)
        snapped_point = route["line"].interpolate(snapped_distance)
        if bus_point.distance(snapped_point) > max_match_distance_m:
            continue
        segment_idx = int(snapped_distance // segment_length_m)
        segment_start = segment_idx * segment_length_m
        segment_end = min(segment_start + segment_length_m, route["line"].length)
        if segment_end - segment_start < 8:
            continue
        segment_geom = substring(route["line"], segment_start, segment_end)
        if segment_geom.is_empty or segment_geom.geom_type not in {"LineString", "MultiLineString"}:
            continue
        segment_id = f"{route_key[0]}-{route_key[1]}-{segment_idx}"
        bucket = aggregates.setdefault(
            segment_id,
            {
                "city": "new_taipei",
                "city_label": "新北市",
                "segment_id": segment_id,
                "section_id": None,
                "section_name": f"{route['route_name']} 第{segment_idx + 1}路段",
                "route_name": route["route_name"],
                "source_type": "公車估計",
                "avg_occ": None,
                "total_vol": None,
                "moe_level": None,
                "segment_length_m": round(float(segment_geom.length), 1),
                "data_time": row.get("UpdateTime"),
                "speed_sum": 0.0,
                "sample_count": 0,
                "geometry": segment_geom,
            },
        )
        bucket["speed_sum"] += float(speed)
        bucket["sample_count"] += 1
        bucket["data_time"] = max(str(bucket["data_time"]), str(row.get("UpdateTime")))

    rows = []
    for bucket in aggregates.values():
        avg_speed = round(bucket["speed_sum"] / bucket["sample_count"], 2)
        status_code, status_label = classify_congestion(avg_speed)
        rows.append(
            {
                k: v
                for k, v in {
                    **bucket,
                    "avg_speed": avg_speed,
                    "status_code": status_code,
                    "status_label": status_label,
                }.items()
                if k != "speed_sum"
            }
        )
    return gpd.GeoDataFrame(rows, geometry="geometry", crs=f"EPSG:{DISTANCE_CRS}")
