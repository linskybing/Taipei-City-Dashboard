from argparse import ArgumentParser
from pathlib import Path
import sys

import pandas as pd
from sqlalchemy import create_engine

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from dags.utils.metrotaipei_parking_supply import write_geojson_snapshots


def parse_args():
    parser = ArgumentParser(
        description="Export metro Taipei parking supply points to FE GeoJSON snapshots."
    )
    parser.add_argument(
        "--dsn",
        required=True,
        help="SQLAlchemy DSN for the ready-data database containing tran_parking_supply_points_metrotaipei.",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    engine = create_engine(args.dsn)
    data = pd.read_sql(
        "SELECT * FROM public.tran_parking_supply_points_metrotaipei", engine
    )
    repo_root = ROOT.parent
    write_geojson_snapshots(data, repo_root)
    print("Exported parking_supply_points_taipei.geojson and parking_supply_points_metrotaipei.geojson")


if __name__ == "__main__":
    main()
