from airflow import DAG
from operators.common_pipeline import CommonDag


def illegal_parking_events_tpe(**kwargs):
    from sqlalchemy import create_engine
    from utils.illegal_parking_events import fetch_taipei_illegal_parking_events
    from utils.load_stage import (
        save_dataframe_to_postgresql,
        update_lasttime_in_data_to_dataset_info,
    )

    dag_infos = kwargs.get("dag_infos")
    ready_data_db_uri = kwargs.get("ready_data_db_uri")
    data = fetch_taipei_illegal_parking_events()
    engine = create_engine(ready_data_db_uri)
    save_dataframe_to_postgresql(
        engine,
        data=data,
        load_behavior=dag_infos.get("load_behavior"),
        default_table=dag_infos.get("ready_data_default_table"),
        history_table=dag_infos.get("ready_data_history_table"),
    )
    update_lasttime_in_data_to_dataset_info(engine, dag_infos.get("dag_id"), data["data_time"].max())


dag = CommonDag(proj_folder="proj_city_dashboard", dag_folder="illegal_parking_events_tpe")
dag.create_dag(etl_func=illegal_parking_events_tpe)
