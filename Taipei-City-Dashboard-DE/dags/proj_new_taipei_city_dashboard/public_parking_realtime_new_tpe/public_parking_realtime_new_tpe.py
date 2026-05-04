from airflow import DAG
from operators.common_pipeline import CommonDag


def public_parking_realtime_new_tpe(**kwargs):
    from sqlalchemy import create_engine
    from utils.load_stage import (
        save_dataframe_to_postgresql,
        update_lasttime_in_data_to_dataset_info,
    )
    from utils.new_taipei_public_parking import fetch_new_taipei_public_parking_realtime

    dag_infos = kwargs.get("dag_infos")
    ready_data = fetch_new_taipei_public_parking_realtime()
    if ready_data.empty:
        raise ValueError("NTPC public parking realtime API returned no rows.")
    engine = create_engine(kwargs.get("ready_data_db_uri"))
    save_dataframe_to_postgresql(
        engine,
        data=ready_data,
        load_behavior=dag_infos.get("load_behavior"),
        default_table=dag_infos.get("ready_data_default_table"),
        history_table=dag_infos.get("ready_data_history_table"),
    )
    update_lasttime_in_data_to_dataset_info(
        engine, dag_infos.get("dag_id"), ready_data["data_time"].max()
    )


dag_builder = CommonDag(
    proj_folder="proj_new_taipei_city_dashboard",
    dag_folder="public_parking_realtime_new_tpe",
)
dag = dag_builder.create_dag(etl_func=public_parking_realtime_new_tpe)
