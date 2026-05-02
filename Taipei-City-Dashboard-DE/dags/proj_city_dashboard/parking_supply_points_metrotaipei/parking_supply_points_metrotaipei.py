from airflow import DAG
from operators.common_pipeline import CommonDag


def parking_supply_points_metrotaipei(**kwargs):
    from sqlalchemy import create_engine
    from utils.load_stage import (
        save_geodataframe_to_postgresql,
        update_lasttime_in_data_to_dataset_info,
    )
    from utils.metrotaipei_parking_supply import build_metrotaipei_parking_supply

    dag_infos = kwargs.get("dag_infos")
    engine = create_engine(kwargs.get("ready_data_db_uri"))
    ready_data = build_metrotaipei_parking_supply(engine)
    if ready_data.empty:
        raise ValueError("Metro Taipei parking supply ETL returned no rows.")
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
    proj_folder="proj_city_dashboard",
    dag_folder="parking_supply_points_metrotaipei",
)
dag = dag_builder.create_dag(etl_func=parking_supply_points_metrotaipei)
