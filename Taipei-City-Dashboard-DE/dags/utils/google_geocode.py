import os
import re
import time
from typing import Optional

import pandas as pd
import requests
from requests.adapters import HTTPAdapter

try:
    from airflow.models import Variable
except Exception:  # pragma: no cover - Airflow is not always available in local tooling
    Variable = None


GOOGLE_GEOCODING_URL = "https://maps.googleapis.com/maps/api/geocode/json"
GOOGLE_KEY_NAMES = ("GOOGLE_GEOCODING_API_KEY", "GOOGLE_MAPS_API_KEY")


def clean_address_for_google(address: str) -> str:
    """Normalize a Taiwan address string for Google Geocoding."""
    if not isinstance(address, str):
        return ""

    trans_table = str.maketrans("０１２３４５６７８９", "0123456789")
    address = address.translate(trans_table).strip()
    address = re.sub(r"\s+", "", address)
    address = re.sub(r"臺灣|台灣|參考地址[:：]", "", address)
    address = address.replace("村里", "里")
    address = address.replace("臺", "台")
    address = re.sub(r"(新北市)+", "新北市", address)
    address = re.sub(r"[一二三四五六七八九十\d]+樓.*$", "", address)
    address = re.sub(r"B\d+.*$", "", address, flags=re.IGNORECASE)
    return address.strip(" ,，")


def _get_google_api_key() -> Optional[str]:
    for key_name in GOOGLE_KEY_NAMES:
        value = os.getenv(key_name)
        if value:
            return value

    if Variable is None:
        return None

    for key_name in GOOGLE_KEY_NAMES:
        try:
            value = Variable.get(key_name)
        except Exception:
            value = None
        if value:
            return value
    return None


def _build_session() -> requests.Session:
    session = requests.Session()
    adapter = HTTPAdapter(max_retries=3)
    session.mount("http://", adapter)
    session.mount("https://", adapter)
    return session


def geocode_address(address: str, api_key: str, session: requests.Session, timeout=30) -> dict:
    """Geocode one address with Google and return lng/lat plus request status."""
    clean_address = clean_address_for_google(address)
    default_result = {
        "query_address": clean_address,
        "matched_address": None,
        "lng": None,
        "lat": None,
        "provider": "google_geocoding",
        "status": "EMPTY_ADDRESS" if not clean_address else "UNKNOWN_ERROR",
    }
    if not clean_address:
        return default_result

    response = session.get(
        GOOGLE_GEOCODING_URL,
        params={
            "address": clean_address,
            "key": api_key,
            "language": "zh-TW",
            "region": "tw",
        },
        timeout=timeout,
    )
    response.raise_for_status()
    data = response.json()
    status = data.get("status", "UNKNOWN_ERROR")
    results = data.get("results", [])
    if status == "OK" and results:
        location = results[0]["geometry"]["location"]
        return {
            "query_address": clean_address,
            "matched_address": results[0].get("formatted_address"),
            "lng": float(location["lng"]),
            "lat": float(location["lat"]),
            "provider": "google_geocoding",
            "status": "OK",
        }
    if status == "ZERO_RESULTS":
        default_result["status"] = "ZERO_RESULTS"
        return default_result

    error_message = data.get("error_message")
    default_result["status"] = f"{status}:{error_message}" if error_message else status
    return default_result


def geocode_address_series(addresses: pd.Series, sleep_seconds=0.1, timeout=30) -> pd.DataFrame:
    """Geocode a series of addresses with deduplication and return a result dataframe."""
    address_list = addresses.fillna("").astype(str).tolist()
    unique_addresses = []
    seen = set()
    for address in address_list:
        clean_address = clean_address_for_google(address)
        if clean_address and clean_address not in seen:
            unique_addresses.append(clean_address)
            seen.add(clean_address)

    api_key = _get_google_api_key()
    if unique_addresses and not api_key:
        raise ValueError(
            "Google geocoding fallback needs GOOGLE_GEOCODING_API_KEY or GOOGLE_MAPS_API_KEY "
            "from environment variables or Airflow Variables."
        )

    session = _build_session()
    cache = {}
    for address in unique_addresses:
        cache[address] = geocode_address(address, api_key=api_key, session=session, timeout=timeout)
        if sleep_seconds:
            time.sleep(sleep_seconds)

    rows = []
    for address in address_list:
        clean_address = clean_address_for_google(address)
        rows.append(cache.get(clean_address, {
            "query_address": clean_address,
            "matched_address": None,
            "lng": None,
            "lat": None,
            "provider": "google_geocoding",
            "status": "EMPTY_ADDRESS",
        }))
    return pd.DataFrame(rows)
