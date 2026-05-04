from airflow import DAG
from operators.common_pipeline import CommonDag


def _traffic_road_congestion_ntpe(**kwargs):
    from sqlalchemy import create_engine
    from utils.extract_stage import get_tdx_data
    from utils.load_stage import save_geodataframe_to_postgresql, update_lasttime_in_data_to_dataset_info
    from utils.traffic_congestion import estimate_new_taipei_segments
    from utils.transform_geometry import convert_geometry_to_wkbgeometry
    from utils.transform_time import convert_str_to_time_format

    ready_data_db_uri = kwargs.get("ready_data_db_uri")
    dag_infos = kwargs.get("dag_infos")
    dag_id = dag_infos.get("dag_id")
    load_behavior = dag_infos.get("load_behavior")
    default_table = dag_infos.get("ready_data_default_table")
    history_table = dag_infos.get("ready_data_history_table")

    realtime_rows = get_tdx_data(
        "https://tdx.transportdata.tw/api/basic/v2/Bus/RealTimeByFrequency/City/NewTaipei?%24format=JSON"
    )
    shape_rows = get_tdx_data(
        "https://tdx.transportdata.tw/api/basic/v2/Bus/Shape/City/NewTaipei?%24format=JSON"
    )
    stop_rows = get_tdx_data(
        "https://tdx.transportdata.tw/api/basic/v2/Bus/StopOfRoute/City/NewTaipei?%24format=JSON"
    )

    gdata = estimate_new_taipei_segments(realtime_rows, shape_rows, stop_rows)
    gdata["data_time"] = convert_str_to_time_format(gdata["data_time"], errors="coerce")
    gdata = convert_geometry_to_wkbgeometry(gdata, from_crs=3826)
    ready_data = gdata[
        [
            "city",
            "city_label",
            "segment_id",
            "section_id",
            "section_name",
            "route_name",
            "source_type",
            "avg_speed",
            "avg_occ",
            "total_vol",
            "sample_count",
            "moe_level",
            "status_code",
            "status_label",
            "segment_length_m",
            "data_time",
            "wkb_geometry",
        ]
    ]

    engine = create_engine(ready_data_db_uri)
    save_geodataframe_to_postgresql(
        engine,
        gdata=ready_data,
        load_behavior=load_behavior,
        default_table=default_table,
        history_table=history_table,
        geometry_type="LineString",
    )
    update_lasttime_in_data_to_dataset_info(engine, dag_id, ready_data["data_time"].max())


dag = CommonDag(proj_folder="proj_new_taipei_city_dashboard", dag_folder="traffic_road_congestion_ntpe")
dag.create_dag(etl_func=_traffic_road_congestion_ntpe)
