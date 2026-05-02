from airflow import DAG
from operators.common_pipeline import CommonDag


def _traffic_road_congestion_tpe(**kwargs):
    from sqlalchemy import create_engine
    from utils.load_stage import save_geodataframe_to_postgresql, update_lasttime_in_data_to_dataset_info
    from utils.traffic_congestion import fetch_taipei_vd_sections
    from utils.transform_geometry import convert_geometry_to_wkbgeometry
    from utils.transform_time import convert_str_to_time_format

    ready_data_db_uri = kwargs.get("ready_data_db_uri")
    dag_infos = kwargs.get("dag_infos")
    dag_id = dag_infos.get("dag_id")
    load_behavior = dag_infos.get("load_behavior")
    default_table = dag_infos.get("ready_data_default_table")
    history_table = dag_infos.get("ready_data_history_table")

    gdata = fetch_taipei_vd_sections("https://tcgbusfs.blob.core.windows.net/blobtisv/GetVD.xml.gz")
    gdata["data_time"] = convert_str_to_time_format(
        gdata["data_time"], from_format="%Y/%m/%dT%H:%M:%S", errors="coerce"
    )
    gdata = convert_geometry_to_wkbgeometry(gdata, from_crs=4326)
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


dag = CommonDag(proj_folder="proj_city_dashboard", dag_folder="traffic_road_congestion_tpe")
dag.create_dag(etl_func=_traffic_road_congestion_tpe)
