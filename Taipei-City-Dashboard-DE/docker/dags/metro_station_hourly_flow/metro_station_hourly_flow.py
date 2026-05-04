from airflow import DAG
from operators.common_pipeline import CommonDag


PAGE_ID = "63f31c7e-7fc3-418b-bd82-b95158755b4d"
MONTHS_TO_LOAD = 3


def _metro_station_hourly_flow(**kwargs):
    import pandas as pd
    from sqlalchemy import create_engine
    from utils.extract_stage import get_current_rid_from_page_id
    from utils.load_stage import (
        save_dataframe_to_postgresql,
        update_lasttime_in_data_to_dataset_info,
    )

    ready_data_db_uri = kwargs.get("ready_data_db_uri")
    dag_infos = kwargs.get("dag_infos")
    dag_id = dag_infos.get("dag_id")
    load_behavior = dag_infos.get("load_behavior")
    default_table = dag_infos.get("ready_data_default_table")
    history_table = dag_infos.get("ready_data_history_table")

    rid = get_current_rid_from_page_id(PAGE_ID)
    index_url = (
        "https://data.taipei/api/frontstage/tpeod/dataset/"
        f"resource.download?rid={rid}"
    )
    index_data = pd.read_csv(index_url)
    index_data["西元年"] = pd.to_numeric(index_data["西元年"], errors="coerce")
    index_data["月"] = pd.to_numeric(index_data["月"], errors="coerce")
    index_data = index_data.dropna(subset=["西元年", "月", "URL"])
    index_data = index_data.sort_values(["西元年", "月"]).tail(MONTHS_TO_LOAD)
    if index_data.empty:
        raise ValueError("No Taipei Metro monthly OD resources found.")

    monthly_frames = [_read_month(row) for _, row in index_data.iterrows()]
    ready_data = pd.concat(monthly_frames, ignore_index=True)
    ready_data = ready_data.sort_values(["data_time", "station_name", "direction"])

    engine = create_engine(ready_data_db_uri)
    save_dataframe_to_postgresql(
        engine,
        data=ready_data,
        load_behavior=load_behavior,
        default_table=default_table,
        history_table=history_table,
    )
    lasttime_in_data = str(ready_data["data_time"].max())
    update_lasttime_in_data_to_dataset_info(engine, dag_id, lasttime_in_data)


def _read_month(row):
    import pandas as pd
    from utils.transform_time import convert_str_to_time_format

    raw_data = pd.read_csv(row["URL"], dtype={"時段": str})
    required = {"日期", "時段", "進站", "出站", "人次"}
    missing = required.difference(raw_data.columns)
    if missing:
        raise ValueError(f"Metro OD file missing columns: {sorted(missing)}")

    data = raw_data.rename(
        columns={
            "日期": "service_date",
            "時段": "hour",
            "進站": "origin_station",
            "出站": "destination_station",
            "人次": "passenger_count",
        }
    )
    data["hour"] = data["hour"].str.zfill(2)
    data["passenger_count"] = pd.to_numeric(
        data["passenger_count"], errors="coerce"
    ).fillna(0)
    data["data_time"] = convert_str_to_time_format(
        data["service_date"] + " " + data["hour"] + ":00:00"
    )

    entries = _aggregate_direction(data, "origin_station", "entry")
    exits = _aggregate_direction(data, "destination_station", "exit")
    month_data = pd.concat([entries, exits], ignore_index=True)
    month_data["source_year"] = int(row["西元年"])
    month_data["source_month"] = int(row["月"])
    month_data["source_seq_no"] = int(row["SeqNo"])
    return month_data


def _aggregate_direction(data, station_col, direction):
    grouped = (
        data.groupby(["data_time", "service_date", "hour", station_col], as_index=False)[
            "passenger_count"
        ]
        .sum()
        .rename(columns={station_col: "station_name"})
    )
    grouped["direction"] = direction
    grouped["passenger_count"] = grouped["passenger_count"].astype(int)
    return grouped[
        [
            "data_time",
            "service_date",
            "hour",
            "station_name",
            "direction",
            "passenger_count",
        ]
    ]


dag = CommonDag(proj_folder="docker_dags", dag_folder="metro_station_hourly_flow")
dag.create_dag(etl_func=_metro_station_hourly_flow)
