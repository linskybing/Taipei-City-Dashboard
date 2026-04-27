-- Fix missing tables for Taipei City Dashboard (Dashboard Data DB)
-- This script initializes the metadata tables required for Data-End ETL flows.
-- Note: Manager DB tables (auth, components, etc.) are handled by Backend AutoMigration.

-- 1. Create trigger function for automatic timestamp updates
CREATE OR REPLACE FUNCTION public.trigger_set_timestamp()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
BEGIN
  NEW._mtime = NOW();
  RETURN NEW;
END;
$function$;

-- 2. Metadata Table (dataset_info)
-- This table is critical for Airflow DAGs to record ingestion status and for the frontend to display data freshness.
CREATE TABLE IF NOT EXISTS public.dataset_info (
    id text PRIMARY KEY,
    psql_table_name text,
    name_cn text,
    airflow_dag_id text,
    mongo_collection text,
    maintain_type text,
    airflow_update_freq text,
    source text,
    source_type text,
    source_department text,
    lasttime_in_data timestamp with time zone,
    resource_updatetime timestamp with time zone,
    gis_format text,
    coordinate text,
    is_geometry boolean,
    dataset_description text,
    etl_description text,
    been_used_count integer DEFAULT 0,
    sensitivity text,
    schedule_interval text,
    _mtime timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    _ctime timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

-- Add trigger for dataset_info to auto-update _mtime
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'dataset_info_mtime') THEN
        CREATE TRIGGER dataset_info_mtime
            BEFORE INSERT OR UPDATE
            ON public.dataset_info
            FOR EACH ROW
            EXECUTE PROCEDURE public.trigger_set_timestamp();
    END IF;
END $$;

-- 3. Sequence Initialization (Optional but recommended for consistency)
-- Ensure standard sequences exist if needed by specific analytical tables.
-- (ETL processes usually create their own sequences or use serial types)

-- 4. Grant Permissions (Ensuring Airflow/Backend can access)
-- Assuming the user 'airflow' or the backend user needs full access to this analytical DB.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'airflow') THEN
        GRANT ALL ON TABLE public.dataset_info TO airflow;
        ALTER TABLE public.dataset_info OWNER TO airflow;
    END IF;
END $$;
