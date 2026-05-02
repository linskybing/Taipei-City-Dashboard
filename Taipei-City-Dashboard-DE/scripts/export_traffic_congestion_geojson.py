import argparse
import json
from io import StringIO
from pathlib import Path
import subprocess

import pandas as pd


ROOT = Path(__file__).resolve().parents[1]
MAP_DIR = ROOT.parent / "Taipei-City-Dashboard-FE" / "public" / "mapData"
SELECT_COLUMNS = """
SELECT
    city,
    city_label,
    segment_id,
    section_id,
    section_name,
    route_name,
    source_type,
    ROUND(avg_speed::numeric, 2) AS avg_speed,
    sample_count::bigint AS sample_count,
    status_code,
    status_label,
    ROUND(NULLIF(segment_length_m::text, '')::numeric, 1) AS segment_length_m,
    data_time::text,
    ST_AsGeoJSON(wkb_geometry::geometry)::text AS geometry_json
"""
TAIPEI_QUERY = f"""
{SELECT_COLUMNS}
FROM public.traffic_road_congestion_tpe
ORDER BY array_position(ARRAY['壅塞','車多','順暢','資料不足'], status_label), avg_speed ASC
"""
METRO_QUERY = f"""
SELECT * FROM (
    {SELECT_COLUMNS}
    FROM public.traffic_road_congestion_tpe
    UNION ALL
    {SELECT_COLUMNS}
    FROM public.traffic_road_congestion_ntpe
) congestion
ORDER BY city, array_position(ARRAY['壅塞','車多','順暢','資料不足'], status_label), avg_speed ASC
"""


def _run_copy_query(container_name, database_name, query):
    sql = f"COPY ({query}) TO STDOUT WITH CSV HEADER"
    result = subprocess.run(
        [
            "docker",
            "exec",
            container_name,
            "psql",
            "-U",
            "postgres",
            "-d",
            database_name,
            "-P",
            "pager=off",
            "-c",
            sql,
        ],
        check=True,
        capture_output=True,
        text=True,
    )
    return pd.read_csv(StringIO(result.stdout))


def _to_geojson(dataframe):
    features = []
    for row in dataframe.to_dict(orient="records"):
        geometry = json.loads(row.pop("geometry_json"))
        for key, value in list(row.items()):
            if pd.isna(value):
                row[key] = None
        features.append({"type": "Feature", "geometry": geometry, "properties": row})
    return {"type": "FeatureCollection", "features": features}


def _write_geojson(payload, output_path):
    output_path.write_text(
        json.dumps(payload, ensure_ascii=False, separators=(",", ":"), allow_nan=False),
        encoding="utf-8",
    )


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--docker-container", default="postgres-data")
    parser.add_argument("--database", default="dashboard")
    args = parser.parse_args()

    MAP_DIR.mkdir(parents=True, exist_ok=True)
    taipei_data = _run_copy_query(args.docker_container, args.database, TAIPEI_QUERY)
    metro_data = _run_copy_query(args.docker_container, args.database, METRO_QUERY)

    _write_geojson(_to_geojson(taipei_data), MAP_DIR / "traffic_congestion_lines_taipei.geojson")
    _write_geojson(_to_geojson(metro_data), MAP_DIR / "traffic_congestion_lines_metrotaipei.geojson")


if __name__ == "__main__":
    main()
