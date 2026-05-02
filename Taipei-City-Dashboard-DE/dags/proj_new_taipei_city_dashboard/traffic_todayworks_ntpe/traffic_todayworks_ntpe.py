from airflow import DAG
from operators.common_pipeline import CommonDag


def _format_roc_series(series):
    values = series.fillna("").astype(str).str.replace(r"\.0$", "", regex=True).str.strip()
    return values.map(lambda value: value.zfill(7) if value.isdigit() else None)


def _build_display_address(row):
    segments = ["新北市", row.get("district") or "", (row.get("village") or "").replace("村里", "里")]
    road_parts = [
        (row.get("road") or "", "路"),
        (row.get("street") or "", "街"),
        (row.get("boulevard") or "", "大道"),
        (row.get("section") or "", "段"),
        (row.get("lane") or "", "巷"),
        (row.get("alley") or "", "弄"),
        (row.get("doorplate") or "", "號"),
    ]
    has_detail = any(value for value, _ in road_parts)
    for value, suffix in road_parts:
        if value:
            segments.append(f"{value}{suffix}")
    structured = "".join(segments)
    if has_detail:
        return structured

    fallback = str(row.get("digsite") or "").strip()
    if fallback and not fallback.startswith("新北市"):
        fallback = f"新北市{fallback}"
    return fallback


def _combine_phone(*parts):
    values = []
    for part in parts:
        if part is None:
            continue
        value = str(part).strip()
        if not value or value.lower() == "nan":
            continue
        values.append(value)
    return " / ".join(values) if values else None


def _traffic_todayworks_ntpe(**kwargs):
    import json

    import pandas as pd
    from sqlalchemy import create_engine
    from utils.extract_stage import NewTaipeiAPIClient
    from utils.get_time import get_tpe_now_time_str
    from utils.google_geocode import clean_address_for_google, geocode_address_series
    from utils.load_stage import (
        save_geodataframe_to_postgresql,
        update_lasttime_in_data_to_dataset_info,
    )
    from utils.transform_geometry import add_point_wkbgeometry_column_to_df, convert_twd97_to_wgs84
    from utils.transform_time import convert_str_to_time_format

    ready_data_db_uri = kwargs.get("ready_data_db_uri")
    dag_infos = kwargs.get("dag_infos")
    dag_id = dag_infos.get("dag_id")
    load_behavior = dag_infos.get("load_behavior")
    default_table = dag_infos.get("ready_data_default_table")
    history_table = dag_infos.get("ready_data_history_table")

    client = NewTaipeiAPIClient("96b6101b-c033-4834-8bd5-e312651db7a0", input_format="json")
    raw_data = pd.DataFrame(client.get_all_data(size=1000))
    raw_data["data_time"] = get_tpe_now_time_str(is_with_tz=True)

    data = raw_data.copy()
    data["caseid"] = pd.to_numeric(data["caseid"], errors="coerce").astype("Int64")
    data["twd97x"] = pd.to_numeric(data["twd97x"], errors="coerce")
    data["twd97y"] = pd.to_numeric(data["twd97y"], errors="coerce")
    data["display_address"] = data.apply(_build_display_address, axis=1)
    data["geocode_query"] = data["display_address"].map(clean_address_for_google)
    data["cb_da"] = _format_roc_series(data["casestartdate_yyymmddroc"])
    data["ce_da"] = _format_roc_series(data["caseenddate_yyymmddroc"])
    data["apptime"] = _format_roc_series(data["examdate_yyymmddroc"])
    data["cb_ad"] = convert_str_to_time_format(data["cb_da"], from_format="%TY%m%d", output_level="date", errors="coerce")
    data["cd_ad"] = convert_str_to_time_format(data["ce_da"], from_format="%TY%m%d", output_level="date", errors="coerce")
    data["apptime_ad"] = convert_str_to_time_format(data["apptime"], from_format="%TY%m%d", output_level="date", errors="coerce")

    data["lng"] = pd.Series([None] * len(data), index=data.index, dtype="object")
    data["lat"] = pd.Series([None] * len(data), index=data.index, dtype="object")
    data["geocode_provider"] = "ntpc_twd97"
    data["geocode_status"] = "SOURCE_TWD97"

    source_mask = data["twd97x"].notna() & data["twd97y"].notna() & data["twd97x"].ne(0) & data["twd97y"].ne(0)
    if source_mask.any():
        data.loc[source_mask, "lng"], data.loc[source_mask, "lat"] = convert_twd97_to_wgs84(
            data.loc[source_mask], "twd97x", "twd97y"
        )

    fallback_mask = data["lng"].isna() | data["lat"].isna()
    if fallback_mask.any():
        geocoded = geocode_address_series(data.loc[fallback_mask, "geocode_query"], sleep_seconds=0.05)
        data.loc[fallback_mask, "lng"] = geocoded["lng"].to_list()
        data.loc[fallback_mask, "lat"] = geocoded["lat"].to_list()
        data.loc[fallback_mask, "geocode_provider"] = geocoded["provider"].to_list()
        data.loc[fallback_mask, "geocode_status"] = geocoded["status"].to_list()

    data["positions"] = data.apply(
        lambda row: json.dumps(
            {
                "district": row.get("district"),
                "village": row.get("village"),
                "road": row.get("road"),
                "street": row.get("street"),
                "boulevard": row.get("boulevard"),
                "section": row.get("section"),
                "lane": row.get("lane"),
                "alley": row.get("alley"),
                "doorplate": row.get("doorplate"),
                "source_address": row.get("digsite"),
                "geocode_query": row.get("geocode_query"),
                "geocode_status": row.get("geocode_status"),
            },
            ensure_ascii=False,
        ),
        axis=1,
    )

    data["ac_no"] = data["caseid"]
    data["st_no"] = None
    data["sno"] = data["licno"]
    data["appmode"] = data["casetype"]
    data["x"] = data["twd97x"]
    data["y"] = data["twd97y"]
    data["app_name"] = data["examunit"]
    data["c_name"] = data["district"]
    data["addr"] = data["display_address"]
    data["co_ti"] = data["workperiod"]
    data["tc_na"] = data["constructionunit"]
    data["tc_ma"] = data["construction_man"]
    data["tc_tl"] = data.apply(
        lambda row: _combine_phone(
            row.get("construction_localcallservice"),
            row.get("construction_tel_ext"),
            row.get("construction_mobiletelephone"),
        ),
        axis=1,
    )
    data["tc_ma3"] = data["supervise_man"]
    data["tc_tl3"] = data.apply(
        lambda row: _combine_phone(
            row.get("supervise_localcallservice"),
            row.get("supervise_tel_ext"),
            row.get("supervise_mobiletelephone"),
        ),
        axis=1,
    )
    data["npurp"] = data["constname"]
    data["dtype"] = data["casetype"]
    data["dlen"] = None
    data["isblock"] = None
    data["isstay"] = None
    data["planb"] = data["remark"]

    gdata = add_point_wkbgeometry_column_to_df(data, x=data["lng"], y=data["lat"], from_crs=4326)
    ready_data = gdata[
        [
            "ac_no", "st_no", "sno", "appmode", "x", "y", "apptime", "app_name", "c_name",
            "addr", "cb_da", "ce_da", "co_ti", "tc_na", "tc_ma", "tc_tl", "tc_ma3", "tc_tl3",
            "npurp", "dtype", "dlen", "positions", "wkb_geometry", "cb_ad", "cd_ad", "apptime_ad",
            "isblock", "isstay", "planb", "lng", "lat", "constname", "digarea", "statdesc",
            "geocode_query", "geocode_provider", "geocode_status", "data_time",
        ]
    ]

    engine = create_engine(ready_data_db_uri)
    save_geodataframe_to_postgresql(
        engine,
        gdata=ready_data,
        load_behavior=load_behavior,
        default_table=default_table,
        history_table=history_table,
        geometry_type="Point",
    )
    update_lasttime_in_data_to_dataset_info(engine, dag_id, ready_data["data_time"].max())


dag = CommonDag(proj_folder="proj_new_taipei_city_dashboard", dag_folder="traffic_todayworks_ntpe")
dag.create_dag(etl_func=_traffic_todayworks_ntpe)
