from airflow import DAG
from operators.common_pipeline import CommonDag


def public_parking_new_tpe(**kwargs):
    from sqlalchemy import create_engine
    from utils.load_stage import (
        save_geodataframe_to_postgresql,
        update_lasttime_in_data_to_dataset_info,
    )
    from utils.new_taipei_public_parking import fetch_new_taipei_public_parking_static
    from utils.transform_geometry import add_point_wkbgeometry_column_to_df

    dag_infos = kwargs.get("dag_infos")
    data = fetch_new_taipei_public_parking_static()
    if data.empty:
        raise ValueError("NTPC public parking static API returned no rows.")
    data = data[(data["x_97"] != 0) & (data["y_97"] != 0)].dropna(subset=["x_97", "y_97"])
    if data.empty:
        raise ValueError("NTPC public parking static API returned no usable coordinates.")
    ready_columns = [
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
        "lng",
        "lat",
        "wkb_geometry",
    ]
    gdata = add_point_wkbgeometry_column_to_df(
        data, data["x_97"], data["y_97"], from_crs=3826
    )
    ready_data = gdata.drop(columns=["geometry", "x_97", "y_97"])[ready_columns]

    engine = create_engine(kwargs.get("ready_data_db_uri"))
    save_geodataframe_to_postgresql(
        engine,
        gdata=ready_data,
        load_behavior=dag_infos.get("load_behavior"),
        default_table=dag_infos.get("ready_data_default_table"),
        history_table=dag_infos.get("ready_data_history_table"),
        geometry_type="Point",
    )
    update_lasttime_in_data_to_dataset_info(
        engine, dag_infos.get("dag_id"), ready_data["data_time"].max()
    )


dag_builder = CommonDag(
    proj_folder="proj_new_taipei_city_dashboard",
    dag_folder="public_parking_new_tpe",
)
dag = dag_builder.create_dag(etl_func=public_parking_new_tpe)
