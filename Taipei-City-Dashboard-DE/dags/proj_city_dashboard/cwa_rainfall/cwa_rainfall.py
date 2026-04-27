from airflow import DAG
from operators.common_pipeline import CommonDag
import pandas as pd
import geopandas as gpd
from shapely.geometry import Point
import requests
from geoalchemy2.shape import from_shape

def _transfer(**kwargs):
    from utils.load_stage import save_geodataframe_to_postgresql, update_lasttime_in_data_to_dataset_info
    from sqlalchemy import create_engine

    # 1. Config
    ready_data_db_uri = kwargs.get('ready_data_db_uri')
    dag_infos = kwargs.get('dag_infos')
    dag_id = dag_infos.get('dag_id')
    load_behavior = dag_infos.get('load_behavior')
    default_table = dag_infos.get('ready_data_default_table')
    history_table = dag_infos.get('ready_data_history_table')
    
    API_URL = "https://opendata.cwa.gov.tw/api//v1/rest/datastore/O-A0002-001?Authorization=CWA-542A2CC7-BDBC-41E9-B191-AA65F66C359A&limit=100&StationId=466850,466881,466900,466910,466920,466930,C0A520,C0A530,C0A550,C0A570,C0A640,C0A770,C0A860,C0A870,C0A890,C0A931,C0A940,C0A950,C0A970,C0A980,C0A9C0,C0A9F0,C0AC40,C0AC60,C0AC70,C0AC80,C0ACA0,C0AD10,C0AD30,C0AD40,C0AD50,C0AG80,C0AH00,C0AH10,C0AH30,C0AH40,C0AH50,C0AH70,C0AH80,C0AH90,C0AI00,C0AI10,C0AI20,C0AI30,C0AI40,C0AJ20,C0AJ30,C0AJ40,C0AJ50,C0AJ60,C0AJ70,C0AJ80,C0AJ90,C0AK10,C0AK30,C1A9N0,C1AC50,C1AI50,C1AI60"

    # 2. Extract
    response = requests.get(API_URL)
    response.raise_for_status()
    raw_data = response.json()
    stations = raw_data.get('records', {}).get('Station', [])

    # 3. Transform
    processed_list = []
    for st in stations:
        # Extract basic info
        obs_time = st.get('ObsTime', {}).get('DateTime')
        geo = st.get('GeoInfo', {})
        rain = st.get('RainfallElement', {})
        
        # Get WGS84 Coordinates
        coords = geo.get('Coordinates', [])
        wgs84 = next((c for c in coords if c.get('CoordinateName') == 'WGS84'), {})
        
        processed_list.append({
            "station_id": st.get('StationId'),
            "station_name": st.get('StationName'),
            "obs_time": obs_time,
            "county": geo.get('CountyName'),
            "town": geo.get('TownName'),
            "lat": float(wgs84.get('StationLatitude')) if wgs84.get('StationLatitude') else None,
            "lon": float(wgs84.get('StationLongitude')) if wgs84.get('StationLongitude') else None,
            "rain_10min": float(rain.get('Past10Min', {}).get('Precipitation', 0)),
            "rain_1hr": float(rain.get('Past1hr', {}).get('Precipitation', 0)),
            "rain_3hr": float(rain.get('Past3hr', {}).get('Precipitation', 0)),
            "rain_24hr": float(rain.get('Past24hr', {}).get('Precipitation', 0)),
            "data_time": obs_time # Using obs_time as data_time
        })

    df = pd.DataFrame(processed_list)
    
    # Clean data: drop rows without coordinates
    df = df.dropna(subset=['lat', 'lon'])
    
    # Create Geometry
    geometry = [Point(xy) for xy in zip(df['lon'], df['lat'])]
    gdf = gpd.GeoDataFrame(df, geometry=geometry, crs="EPSG:4326")
    
    # Rename for standard consistency
    gdf = gdf.rename(columns={'geometry': 'wkb_geometry'})
    # Convert to WKBElement for GeoAlchemy2 compatibility
    gdf['wkb_geometry'] = gdf['wkb_geometry'].apply(lambda x: from_shape(x, srid=4326))

    # 4. Load
    engine = create_engine(ready_data_db_uri)
    save_geodataframe_to_postgresql(
        engine, gdata=gdf, load_behavior=load_behavior,
        geometry_type='Point', default_table=default_table,
        history_table=history_table, geometry_col='wkb_geometry'
    )
    
    # Update lasttime_in_data
    if not gdf.empty:
        lasttime_in_data = gdf['data_time'].max()
        update_lasttime_in_data_to_dataset_info(
            engine, airflow_dag_id=dag_id, lasttime_in_data=lasttime_in_data
        )

# Create the DAG
dag = CommonDag(proj_folder='proj_city_dashboard', dag_folder='cwa_rainfall')
dag.create_dag(etl_func=_transfer)
