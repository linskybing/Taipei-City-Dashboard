import argparse
from pathlib import Path

import psycopg2


ROOT = Path(__file__).resolve().parents[1]
SQL_PATH = ROOT.parent / "db-sample-data" / "dashboardmanager-traffic-construction.sql"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--dsn",
        default="postgresql://postgres:hothothot@localhost:5432/dashboardmanager",
        help="PostgreSQL DSN for dashboardmanager",
    )
    args = parser.parse_args()

    sql_text = SQL_PATH.read_text(encoding="utf-8")
    conn = psycopg2.connect(args.dsn)
    try:
        conn.set_client_encoding("UTF8")
        with conn:
            with conn.cursor() as cur:
                cur.execute(sql_text)
    finally:
        conn.close()


if __name__ == "__main__":
    main()
