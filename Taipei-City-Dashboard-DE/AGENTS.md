## Data-End Agent Rules

These rules inherit the root hackathon redlines. If this file conflicts with the root `AGENTS.md`, the root hackathon rules win.

## Scope

- Work in Airflow DAGs, `CommonDag`, Python ETL utilities, `job_config.json`, PostgreSQL dashboard tables, and Docker-based local Airflow setup.
- Read `taipei-city-dashboard-data-end` before editing DAGs, metadata, open-data ingestion, transforms, loads, or Docker settings.

## Compliance Boundaries

- Use legal, transparent open data sources with public benefit. Prefer `data.taipei` and `data.ntpc.gov.tw`; record source proof in metadata.
- Prefer UTF-8 CSV imports when preparing local data.
- Do not over-shape Data-End tables for one frontend component. Store reusable, source-oriented standardized data and let backend queries adapt it.
- Do not use excessive mock data for competition deliverables.
- Do not run database initialization, destructive cleanup, or reset commands unless the user explicitly asks.

## Implementation Standards

- Use `CommonDag` for new DAGs unless maintaining an existing pattern requires otherwise.
- Keep `job_config.json` complete: source platform, authority, URL, transfer format, data range, maintenance strategy, update frequency, and DB table name.
- Use source-oriented snake_case table names with department/source context.
- Declare the maintenance strategy explicitly: `append`, `replace`, or `current+history`.
- Preserve standard fields where applicable: `data_time`, `wkb_geometry` in EPSG:4326, `_ctime`, `_mtime`, and stable/generated keys.
- Run targeted DAG/unit tests only; validate Airflow and tables manually when the user has enabled the local environment.
