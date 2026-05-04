import json
import os
from datetime import datetime
from pathlib import Path

from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.providers.postgres.hooks.postgres import PostgresHook
from sqlalchemy import create_engine, text
from sqlalchemy.engine import make_url


PAYLOAD_PATH = Path(__file__).resolve().with_name("component_seed.json")


def _ensure_component_map(conn, item):
    existing = conn.execute(
        text('SELECT id FROM public.component_maps WHERE "index" = :index AND title = :title'),
        {"index": item["index"], "title": item["title"]},
    ).scalar()
    payload = {
        "index": item["index"],
        "title": item["title"],
        "type": item["type"],
        "source": item["source"],
        "size": item.get("size"),
        "icon": item.get("icon"),
        "paint": json.dumps(item["paint"], ensure_ascii=False),
        "property": json.dumps(item["property"], ensure_ascii=False),
    }
    if existing:
        conn.execute(
            text(
                """
                UPDATE public.component_maps
                SET type = :type, source = :source, size = :size, icon = :icon,
                    paint = CAST(:paint AS json), property = CAST(:property AS json)
                WHERE id = :id
                """
            ),
            {**payload, "id": existing},
        )
        return existing
    return conn.execute(
        text(
            """
            INSERT INTO public.component_maps ("index", title, type, source, size, icon, paint, property)
            VALUES (:index, :title, :type, :source, :size, :icon, CAST(:paint AS json), CAST(:property AS json))
            RETURNING id
            """
        ),
        payload,
    ).scalar_one()


def _ensure_component(conn, item):
    return conn.execute(
        text(
            """
            INSERT INTO public.components ("index", name)
            VALUES (:index, :name)
            ON CONFLICT ("index") DO UPDATE SET name = EXCLUDED.name
            RETURNING id
            """
        ),
        item,
    ).scalar_one()


def _get_component_ids(conn, component_ids, external_indices):
    ids = list(component_ids)
    if not external_indices:
        return ids
    rows = conn.execute(
        text('SELECT "index", id FROM public.components WHERE "index" = ANY(:indices)'),
        {"indices": external_indices},
    ).mappings()
    external_ids = {row["index"]: row["id"] for row in rows}
    missing = [index for index in external_indices if index not in external_ids]
    if missing:
        print(f"[parking_components_metrotaipei] Skip missing: {missing}", flush=True)
    ids.extend(external_ids[index] for index in external_indices if index in external_ids)
    return ids


def _replace_query_chart(conn, item, map_config_ids):
    conn.execute(
        text('DELETE FROM public.query_charts WHERE "index" = :index AND city = :city'),
        {"index": item["index"], "city": item["city"]},
    )
    conn.execute(
        text(
            """
            INSERT INTO public.query_charts (
              "index", history_config, map_config_ids, map_filter, time_from, time_to,
              update_freq, update_freq_unit, source, short_desc, long_desc, use_case,
              links, contributors, created_at, updated_at, query_type, query_chart, query_history, city
            ) VALUES (
              :index, NULL, :map_config_ids, :map_filter, :time_from, :time_to,
              :update_freq, :update_freq_unit, :source, :short_desc, :long_desc, :use_case,
              :links, :contributors, NOW(), NOW(), :query_type, :query_chart, NULL, :city
            )
            """
        ),
        {
            **item,
            "map_config_ids": map_config_ids,
            "map_filter": json.dumps(item["map_filter"], ensure_ascii=False)
            if item["map_filter"]
            else None,
        },
    )


def seed_parking_components_metrotaipei(**_kwargs):
    payload = json.loads(PAYLOAD_PATH.read_text(encoding="utf-8"))
    conn_id = "dashboad-postgre"
    try:
        hook = PostgresHook(postgres_conn_id=conn_id)
        uri = hook.get_uri()
    except Exception as exc:
        conn_id = "postgres_manager_fallback"
        base_uri = PostgresHook(postgres_conn_id="postgres_default").get_uri()
        manager_url = make_url(base_uri).set(
            host=os.getenv("DB_MANAGER_HOST", "postgres-manager"),
            port=int(os.getenv("DB_MANAGER_PORT", "5432")),
            database=os.getenv("DB_MANAGER_DBNAME", "dashboardmanager"),
        )
        uri = str(manager_url)
        print(
            f"[parking_components_metrotaipei] Fallback to {conn_id}: {exc}",
            flush=True,
        )
    engine = create_engine(uri)
    with engine.begin() as conn:
        map_ids = {item["key"]: _ensure_component_map(conn, item) for item in payload["component_maps"]}
        component_ids = {item["index"]: _ensure_component(conn, item) for item in payload["components"]}
        for chart in payload["component_charts"]:
            conn.execute(
                text(
                    """
                    INSERT INTO public.component_charts ("index", color, "types", unit)
                    VALUES (:index, :color, :types, :unit)
                    ON CONFLICT ("index") DO UPDATE
                    SET color = EXCLUDED.color, "types" = EXCLUDED."types", unit = EXCLUDED.unit
                    """
                ),
                chart,
            )
        for query in payload["query_charts"]:
            _replace_query_chart(conn, query, [map_ids[key] for key in query["map_config_keys"]])
        for dashboard in payload["dashboards"]:
            local_ids = [component_ids[index] for index in dashboard["component_indices"]]
            dashboard_component_ids = _get_component_ids(
                conn, local_ids, dashboard.get("external_component_indices", [])
            )
            dashboard_id = conn.execute(
                text(
                    """
                    INSERT INTO public.dashboards ("index", name, components, icon, updated_at, created_at)
                    VALUES (:index, :name, :components, :icon, NOW(), NOW())
                    ON CONFLICT ("index") DO UPDATE
                    SET name = EXCLUDED.name, components = EXCLUDED.components, icon = EXCLUDED.icon, updated_at = NOW()
                    RETURNING id
                    """
                ),
                {
                    "index": dashboard["index"],
                    "name": dashboard["name"],
                    "icon": dashboard["icon"],
                    "components": dashboard_component_ids,
                },
            ).scalar_one()
            conn.execute(
                text("DELETE FROM public.dashboard_groups WHERE dashboard_id = :dashboard_id"),
                {"dashboard_id": dashboard_id},
            )
            conn.execute(
                text(
                    "INSERT INTO public.dashboard_groups (dashboard_id, group_id) VALUES (:dashboard_id, :group_id)"
                ),
                {"dashboard_id": dashboard_id, "group_id": dashboard["group_id"]},
            )


dag = DAG(
    dag_id="parking_components_metrotaipei",
    start_date=datetime(2026, 5, 2),
    schedule=None,
    catchup=False,
    tags=["dashboardmanager", "雙北停車", "component-seed"],
    description="Seed Taipei and Metro Taipei parking supply and price components.",
)

PythonOperator(
    task_id="seed_parking_components_metrotaipei",
    python_callable=seed_parking_components_metrotaipei,
    dag=dag,
)
