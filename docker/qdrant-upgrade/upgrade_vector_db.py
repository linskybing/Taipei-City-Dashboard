import os
import sys
import time
import uuid
import warnings
from urllib.parse import quote_plus

os.environ.setdefault("HF_HUB_DISABLE_XET", "1")
os.environ.setdefault("TOKENIZERS_PARALLELISM", "false")
warnings.filterwarnings("ignore", category=DeprecationWarning)

import pandas as pd
from qdrant_client import QdrantClient
from qdrant_client.http.models import Distance, PointStruct, VectorParams
from sentence_transformers import SentenceTransformer
from sqlalchemy import create_engine


def log(message):
    print(message, flush=True)


def require_env(name):
    value = os.getenv(name)
    if not value:
        raise RuntimeError(f"缺少必要環境變數：{name}")
    return value


def get_collection_name():
    return os.getenv("QDRANT_COLLECTION_NAME") or os.getenv("QDRANT_COLLECTION") or "query_charts"


def get_int_env(name, default):
    raw = os.getenv(name)
    if not raw:
        return default
    try:
        value = int(raw)
    except ValueError as exc:
        raise RuntimeError(f"{name} 必須是整數，目前是 {raw!r}") from exc
    return value if value > 0 else default


def load_source_data(engine):
    query = """
        SELECT c.id, qc.index, c.name, qc.city, qc.long_desc, qc.use_case
        FROM query_charts qc
        INNER JOIN components c ON qc.index = c.index
        WHERE c.id IN (
            SELECT DISTINCT unnest(components)
            FROM dashboards d
            WHERE d.id IN (
                SELECT DISTINCT dashboard_id
                FROM dashboard_groups dg
                WHERE group_id IN (
                    SELECT DISTINCT id FROM "groups" g WHERE is_personal IS false
                )
            )
        )
    """
    return pd.read_sql(query, engine)


def prepare_text(df):
    df["long_desc"] = df["long_desc"].fillna("")
    df["use_case"] = df["use_case"].fillna("")
    df["name"] = df["name"].fillna("")
    df["text"] = (
        "passage: "
        + df["name"].astype(str)
        + " "
        + df["long_desc"].astype(str)
        + " "
        + df["use_case"].astype(str)
    )
    return df


def make_point(row, vector):
    key = f"{row['id']}:{row['index']}:{row['city']}"
    point_id = str(uuid.uuid5(uuid.NAMESPACE_URL, key))
    return PointStruct(
        id=point_id,
        vector=vector.tolist(),
        payload={
            "id": row["id"],
            "index": row["index"],
            "name": row["name"],
            "city": row["city"],
            "long_desc": row["long_desc"],
            "use_case": row["use_case"],
        },
    )


def recreate_collection(client, collection_name, vector_size):
    collections = client.get_collections().collections
    if any(c.name == collection_name for c in collections):
        log(f"刪除舊 collection：{collection_name}")
        client.delete_collection(collection_name)
    log(f"建立新 collection：{collection_name}，vector size={vector_size}")
    client.recreate_collection(
        collection_name=collection_name,
        vectors_config=VectorParams(size=vector_size, distance=Distance.COSINE),
    )


def encode(model, texts, batch_size):
    return model.encode(
        texts,
        batch_size=batch_size,
        normalize_embeddings=True,
        show_progress_bar=False,
    )


def upsert_batches(client, collection_name, model, df, batch_size):
    records = df.to_dict(orient="records")
    texts = df["text"].tolist()
    if not records:
        log("沒有公開 dashboard component 資料，略過 Qdrant 寫入。")
        return 0

    log("生成第一筆向量以判斷 vector size...")
    first_vector = encode(model, [texts[0]], batch_size)[0]
    recreate_collection(client, collection_name, len(first_vector))

    uploaded = 0
    total = len(records)
    for start in range(0, total, batch_size):
        end = min(start + batch_size, total)
        log(f"生成向量並上傳 batch {start + 1}-{end}/{total}...")
        vectors = encode(model, texts[start:end], batch_size)
        points = [make_point(row, vector) for vector, row in zip(vectors, records[start:end])]
        client.upsert(collection_name=collection_name, points=points, wait=True)
        uploaded += len(points)
        log(f"已上傳 {uploaded}/{total} 筆向量至 Qdrant")
    return uploaded


def build_db_engine(host, port, user, password, db_name):
    encoded_user = quote_plus(user)
    encoded_password = quote_plus(password)
    conn = f"postgresql://{encoded_user}:{encoded_password}@{host}:{port}/{db_name}"
    return create_engine(conn)


def main():
    started_at = time.time()
    log("開始向量資料庫升級...")
    db_host = os.getenv("DB_MANAGER_HOST", "postgres-manager")
    db_port = os.getenv("DB_MANAGER_PORT", "5432")
    db_user = require_env("DB_MANAGER_USER")
    db_password = require_env("DB_MANAGER_PASSWORD")
    db_name = require_env("DB_MANAGER_DBNAME")
    qdrant_url = os.getenv("QDRANT_URL", "http://qdrant:6333")
    qdrant_api_key = os.getenv("QDRANT_API_KEY", "gogosecurity")
    collection_name = get_collection_name()
    model_name = os.getenv("EMBEDDING_MODEL_NAME", "intfloat/multilingual-e5-base")
    model_cache = os.getenv("SENTENCE_TRANSFORMERS_HOME")
    batch_size = get_int_env("VECTOR_BATCH_SIZE", 16)

    log(f"目標資料庫：{db_host}:{db_port}/{db_name}")
    log(f"Qdrant URL：{qdrant_url}")
    log(f"Qdrant collection：{collection_name}")
    log(f"Embedding model：{model_name}")
    log(f"Vector batch size：{batch_size}")
    if model_cache:
        log(f"SentenceTransformers cache：{model_cache}")

    log("連接 PostgreSQL...")
    engine = build_db_engine(db_host, db_port, db_user, db_password, db_name)
    log("執行 SQL 查詢並讀取公開 dashboard component 資料...")
    df = prepare_text(load_source_data(engine))
    log(f"成功讀取 {len(df)} 筆資料")

    log("載入 SentenceTransformer 模型...")
    model = SentenceTransformer(model_name, cache_folder=model_cache)
    log("模型載入完成")

    log("連線到 Qdrant...")
    client = QdrantClient(url=qdrant_url, api_key=qdrant_api_key)
    uploaded = upsert_batches(client, collection_name, model, df, batch_size)
    log(f"成功上傳 {uploaded} 筆向量至 Qdrant")
    log(f"向量資料庫升級完成！耗時 {time.time() - started_at:.1f} 秒")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        log(f"向量資料庫升級失敗：{exc}")
        sys.exit(1)
