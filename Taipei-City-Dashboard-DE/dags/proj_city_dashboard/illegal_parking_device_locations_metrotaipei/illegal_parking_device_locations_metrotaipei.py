from airflow import DAG
from operators.common_pipeline import CommonDag


def illegal_parking_device_locations_metrotaipei(**kwargs):
    from sqlalchemy import create_engine
    from utils.illegal_parking_devices import fetch_metrotaipei_illegal_parking_devices
    from utils.load_stage import (
        save_geodataframe_to_postgresql,
        update_lasttime_in_data_to_dataset_info,
    )
    from utils.transform_geometry import add_point_wkbgeometry_column_to_df

    dag_infos = kwargs.get("dag_infos")
    ready_data_db_uri = kwargs.get("ready_data_db_uri")
    data = fetch_metrotaipei_illegal_parking_devices()
    gdata = add_point_wkbgeometry_column_to_df(data=data, x=data["lng"], y=data["lat"], from_crs=4326)
    ready_data = gdata[
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
            "wkb_geometry",
        ]
    ]
    engine = create_engine(ready_data_db_uri)
    save_geodataframe_to_postgresql(
        engine,
        gdata=ready_data,
        load_behavior=dag_infos.get("load_behavior"),
        default_table=dag_infos.get("ready_data_default_table"),
        history_table=dag_infos.get("ready_data_history_table"),
        geometry_type="Point",
    )
    update_lasttime_in_data_to_dataset_info(engine, dag_infos.get("dag_id"), ready_data["data_time"].max())


dag = CommonDag(
    proj_folder="proj_city_dashboard",
    dag_folder="illegal_parking_device_locations_metrotaipei",
)
dag.create_dag(etl_func=illegal_parking_device_locations_metrotaipei)
