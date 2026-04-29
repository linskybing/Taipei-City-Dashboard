import json
import os
import re
import ssl
from html import unescape
from pathlib import Path
from urllib.parse import urlencode
from urllib.request import Request, urlopen

from airflow import DAG
from operators.common_pipeline import CommonDag


def _request_json(url, data=None, timeout=30, verify_ssl=True):
    headers = {
        "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
        "Referer": "https://wellbeing.mohw.gov.tw/nor/mmap/",
        "User-Agent": "Taipei-City-Dashboard-DE/1.0",
        "X-Requested-With": "XMLHttpRequest",
    }
    body = urlencode(data).encode("utf-8") if data else None
    req = Request(url, data=body, headers=headers, method="POST" if data else "GET")
    context = None if verify_ssl else ssl._create_unverified_context()
    with urlopen(req, timeout=timeout, context=context) as res:
        return json.loads(res.read().decode("utf-8"))


def _flatten_districts(area_payload, county_names):
    districts = []
    areas = area_payload.get("data", {})
    for county in areas.values():
        if county.get("name") not in county_names:
            continue
        for district in county.get("childs", {}).values():
            districts.append({
                "county": county.get("name"),
                "district": district.get("name"),
                "area_id": district.get("areaID"),
            })
    return districts


def _clean_text(value):
    if value is None:
        return ""
    text = re.sub(r"<[^>]+>", " ", str(value))
    text = unescape(text).replace("\xa0", " ")
    return re.sub(r"\s+", " ", text).strip()


def _normalize_url(value):
    value = _clean_text(value)
    return "" if value in {"", "無", "none", "None", "null"} else value


def _extract_sites(base_url, districts, tag_type, timeout, verify_ssl):
    rows = []
    for district in districts:
        payload = {
            "areaID": district["area_id"],
            "keyword": "",
            "tagtype": tag_type,
        }
        data = _request_json(f"{base_url}/get_list", payload, timeout, verify_ssl)
        if not data.get("status"):
            continue
        for item in data.get("data", []):
            item["county"] = district["county"]
            item["district"] = district["district"]
            rows.append(item)
    return rows


def _build_geojson(df, counties):
    features = []
    subset = df[df["county"].isin(counties)]
    for row in subset.to_dict("records"):
        props = {
            k: _to_json_value(v)
            for k, v in row.items()
            if k not in {"lng", "lat", "wkb_geometry"}
        }
        features.append({
            "type": "Feature",
            "geometry": {"type": "Point", "coordinates": [row["lng"], row["lat"]]},
            "properties": props,
        })
    return {"type": "FeatureCollection", "features": features}


def _to_json_value(value):
    if value is None:
        return None
    if hasattr(value, "isoformat"):
        return value.isoformat()
    if value != value:
        return None
    return value


def _export_geojson(df, outputs, output_dir):
    if not output_dir:
        print("GeoJSON output skipped: geojson_output_dir is not configured.")
        return
    output_root = Path(output_dir)
    if not output_root.exists() or not os.access(output_root, os.W_OK):
        print(f"GeoJSON output skipped: {output_root} is not writable.")
        return
    for config in outputs:
        target = output_root / config["filename"]
        geojson = _build_geojson(df, config["county_names"])
        target.write_text(json.dumps(geojson, ensure_ascii=False), encoding="utf-8")


def _transfer(**kwargs):
    import pandas as pd
    from sqlalchemy import create_engine
    from utils.get_time import get_tpe_now_time_str
    from utils.load_stage import (
        save_geodataframe_to_postgresql,
        update_lasttime_in_data_to_dataset_info,
    )
    from utils.transform_geometry import add_point_wkbgeometry_column_to_df

    dag_infos = kwargs.get("dag_infos")
    base_url = dag_infos.get("api_base_url")
    county_names = dag_infos.get("county_names")
    tag_type = dag_infos.get("tag_type")
    timeout = dag_infos.get("request_timeout", 30)
    verify_ssl = dag_infos.get("verify_ssl", True)
    output_dir = (
        dag_infos.get("geojson_output_dir") or
        os.getenv("DASHBOARD_FE_MAPDATA_DIR", "")
    )

    areas = _request_json(
        f"{base_url}/get_areas", timeout=timeout, verify_ssl=verify_ssl
    )
    districts = _flatten_districts(areas, county_names)
    raw_data = pd.DataFrame(
        _extract_sites(base_url, districts, tag_type, timeout, verify_ssl)
    )
    raw_data = raw_data.drop_duplicates(subset=["organizeID"]).reset_index(drop=True)

    data = raw_data.rename(columns={
        "organizeID": "organize_id",
        "areaID": "area_id",
        "tag_type": "site_type",
    })
    data["site_name"] = data["title"].map(_clean_text)
    data["site_type"] = data["site_type"].fillna("未分類").map(_clean_text)
    data["address"] = data["address"].map(_clean_text)
    data["tel"] = data["tel"].map(_clean_text)
    data["url"] = data["url"].map(_normalize_url)
    data["time_info"] = data["time_info"].map(_clean_text)
    data["area_label"] = data["county"] + " " + data["district"]
    data["lng"] = pd.to_numeric(data["lon"], errors="coerce")
    data["lat"] = pd.to_numeric(data["lat"], errors="coerce")
    data["created_at_source"] = pd.to_datetime(data["create_time"], errors="coerce")
    data["updated_at_source"] = pd.to_datetime(data["update_time"], errors="coerce")
    data["data_time"] = get_tpe_now_time_str(is_with_tz=True)
    data = data[data["lng"].between(119, 123) & data["lat"].between(21, 26.5)]

    fields = [
        "organize_id", "code", "site_name", "county", "district",
        "area_label", "area_id", "site_type", "address", "tel", "url",
        "time_info", "lng", "lat", "created_at_source",
        "updated_at_source", "data_time",
    ]
    gdata = add_point_wkbgeometry_column_to_df(
        data[fields], data["lng"], data["lat"], from_crs=4326
    )
    ready_data = gdata[fields + ["wkb_geometry"]]

    engine = create_engine(kwargs.get("ready_data_db_uri"))
    save_geodataframe_to_postgresql(
        engine,
        gdata=ready_data,
        load_behavior=dag_infos.get("load_behavior"),
        default_table=dag_infos.get("ready_data_default_table"),
        history_table=dag_infos.get("ready_data_history_table"),
        geometry_type="Point",
    )
    _export_geojson(ready_data, dag_infos.get("geojson_outputs", []), output_dir)
    update_lasttime_in_data_to_dataset_info(
        engine, dag_infos.get("dag_id"), ready_data["data_time"].max()
    )


dag = CommonDag(
    proj_folder="proj_city_dashboard",
    dag_folder="mental_health_support_sites",
)
dag.create_dag(etl_func=_transfer)
