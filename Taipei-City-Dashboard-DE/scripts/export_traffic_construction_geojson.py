import argparse
import json
from io import StringIO
from pathlib import Path
import subprocess

import pandas as pd

ROOT = Path(__file__).resolve().parents[1]
MAP_DIR = ROOT.parent / "Taipei-City-Dashboard-FE" / "public" / "mapData"
DISTRICT_SOURCE = MAP_DIR / "metrotaipei_town.geojson"
ACTIVE_FILTER = "(cb_ad IS NULL OR cb_ad <= CURRENT_DATE) AND (cd_ad IS NULL OR cd_ad >= CURRENT_DATE)"
SITE_QUERY = f"""
WITH sites AS (
    SELECT
        ac_no::text AS permit_id,
        'taipei'::text AS city,
        '臺北市'::text AS city_label,
        c_name::text AS district,
        c_name::text AS district_display,
        addr,
        npurp,
        app_name,
        tc_na,
        co_ti,
        to_char(cb_ad, 'YYYY-MM-DD') AS start_date,
        to_char(cd_ad, 'YYYY-MM-DD') AS end_date,
        '施工中'::text AS status_label,
        ST_X(wkb_geometry::geometry) AS lng,
        ST_Y(wkb_geometry::geometry) AS lat
    FROM public.traffic_todayworks
    WHERE {ACTIVE_FILTER}
    UNION ALL
    SELECT
        ac_no::text AS permit_id,
        'new_taipei'::text AS city,
        '新北市'::text AS city_label,
        c_name::text AS district,
        c_name::text AS district_display,
        addr,
        npurp,
        app_name,
        tc_na,
        co_ti,
        to_char(cb_ad, 'YYYY-MM-DD') AS start_date,
        to_char(cd_ad, 'YYYY-MM-DD') AS end_date,
        COALESCE(NULLIF(statdesc, ''), '施工中') AS status_label,
        ST_X(wkb_geometry::geometry) AS lng,
        ST_Y(wkb_geometry::geometry) AS lat
    FROM public.traffic_todayworks_ntpe
    WHERE {ACTIVE_FILTER}
)
SELECT * FROM sites ORDER BY city, district, start_date NULLS LAST, permit_id
"""


def _normalize_district(city_label, district_name):
    district = str(district_name or "").strip()
    if city_label == "臺北市" and district.endswith("區"):
        return district[:-1]
    return district


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
        lng = row.pop("lng", None)
        lat = row.pop("lat", None)
        if pd.isna(lng) or pd.isna(lat):
            continue
        for key, value in list(row.items()):
            if pd.isna(value):
                row[key] = None
        features.append(
            {
                "type": "Feature",
                "geometry": {"type": "Point", "coordinates": [float(lng), float(lat)]},
                "properties": row,
            }
        )
    return {"type": "FeatureCollection", "features": features}


def _write_geojson(payload, output_path):
    output_path.write_text(
        json.dumps(payload, ensure_ascii=False, separators=(",", ":"), allow_nan=False),
        encoding="utf-8",
    )


def _build_district_geojson(source_data, counts, city_filter=None):
    features = []
    for feature in source_data.get("features", []):
        props = feature.get("properties", {})
        city_label = props.get("PNAME")
        district_display = props.get("TNAME")
        if not city_label or not district_display:
            continue
        if city_filter and city_label != city_filter:
            continue
        district = _normalize_district(city_label, district_display)
        features.append(
            {
                "type": "Feature",
                "geometry": feature.get("geometry"),
                "properties": {
                    "city_label": city_label,
                    "district": district,
                    "district_display": district_display,
                    "name": f"{city_label}{district_display}",
                    "active_count": int(counts.get((city_label, district), 0)),
                },
            }
        )
    return {"type": "FeatureCollection", "features": features}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--docker-container", default="postgres-data")
    parser.add_argument("--database", default="dashboard")
    args = parser.parse_args()

    MAP_DIR.mkdir(parents=True, exist_ok=True)
    site_data = _run_copy_query(args.docker_container, args.database, SITE_QUERY)
    site_data["district"] = [
        _normalize_district(city_label, district)
        for city_label, district in zip(site_data["city_label"], site_data["district"])
    ]
    site_data["district_display"] = [
        district if city != "taipei" or str(district).endswith("區") else f"{district}區"
        for city, district in zip(site_data["city"], site_data["district"])
    ]

    counts = {
        (row["city_label"], row["district"]): int(row["active_count"])
        for row in site_data.groupby(["city_label", "district"]).size().reset_index(name="active_count").to_dict(orient="records")
    }
    district_source = json.loads(DISTRICT_SOURCE.read_text(encoding="utf-8"))

    _write_geojson(
        _build_district_geojson(district_source, counts, city_filter="臺北市"),
        MAP_DIR / "traffic_construction_district_taipei.geojson",
    )
    _write_geojson(
        _build_district_geojson(district_source, counts),
        MAP_DIR / "traffic_construction_district_metrotaipei.geojson",
    )
    _write_geojson(
        _to_geojson(site_data[site_data["city"] == "taipei"].copy()),
        MAP_DIR / "traffic_construction_sites_taipei.geojson",
    )
    _write_geojson(_to_geojson(site_data.copy()), MAP_DIR / "traffic_construction_sites_metrotaipei.geojson")


if __name__ == "__main__":
    main()
