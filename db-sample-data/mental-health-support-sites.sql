BEGIN;

DELETE FROM public.query_charts WHERE index = 'mental_health_support_sites';
DELETE FROM public.component_charts WHERE index = 'mental_health_support_sites';
DELETE FROM public.components WHERE index = 'mental_health_support_sites';
DELETE FROM public.component_maps WHERE id IN (9201, 9202);

INSERT INTO public.components (id, index, name)
VALUES (9201, 'mental_health_support_sites', '心理健康支持機構數量');

INSERT INTO public.component_charts (index, color, types, unit)
VALUES (
  'mental_health_support_sites',
  ARRAY['#2F80ED', '#27AE60', '#F2C94C', '#EB5757', '#9B51E0', '#56CCF2'],
  ARRAY['BarChart', 'ColumnChart'],
  '處'
);

INSERT INTO public.component_maps
  (id, index, title, type, source, size, icon, paint, property)
VALUES
  (
    9201,
    'mental_health_support_sites_tpe',
    '心理健康支持機構',
    'circle',
    'geojson',
    'small',
    NULL,
    '{"circle-color":["match",["get","site_type"],"心理機構","#2F80ED","社區心理衛生中心","#27AE60","精神科醫療機構","#EB5757","精神復健機構","#F2C94C","精神護理機構","#9B51E0","#56CCF2"],"circle-opacity":0.82,"circle-stroke-color":"#FFFFFF","circle-stroke-width":1}',
    '[{"key":"site_name","name":"機構名稱"},{"key":"site_type","name":"類型"},{"key":"county","name":"縣市"},{"key":"district","name":"行政區"},{"key":"address","name":"地址"},{"key":"tel","name":"電話"},{"key":"time_info","name":"服務資訊"},{"key":"url","name":"網站"}]'
  ),
  (
    9202,
    'mental_health_support_sites_metrotaipei',
    '心理健康支持機構',
    'circle',
    'geojson',
    'small',
    NULL,
    '{"circle-color":["match",["get","site_type"],"心理機構","#2F80ED","社區心理衛生中心","#27AE60","精神科醫療機構","#EB5757","精神復健機構","#F2C94C","精神護理機構","#9B51E0","#56CCF2"],"circle-opacity":0.82,"circle-stroke-color":"#FFFFFF","circle-stroke-width":1}',
    '[{"key":"site_name","name":"機構名稱"},{"key":"site_type","name":"類型"},{"key":"county","name":"縣市"},{"key":"district","name":"行政區"},{"key":"address","name":"地址"},{"key":"tel","name":"電話"},{"key":"time_info","name":"服務資訊"},{"key":"url","name":"網站"}]'
  );

INSERT INTO public.query_charts (
  index, history_config, map_config_ids, map_filter, time_from, time_to,
  update_freq, update_freq_unit, source, short_desc, long_desc, use_case,
  links, contributors, created_at, updated_at, query_type, query_chart,
  query_history, city
) VALUES
  (
    'mental_health_support_sites', NULL, ARRAY[9201],
    '{"mode":"byParam","byParam":{"xParam":"district"}}',
    'current', NULL,
    1, 'day', '衛生福利部',
    '顯示臺北市心理健康支持機構數量。',
    '此元件統計臺北市各行政區心理健康支持機構數量，並可在地圖檢視機構名稱、類型、地址、電話、服務資訊與網站。',
    '可用於學生與青壯世代心理健康支持資源盤點，協助找出機構集中或資源較少的行政區。',
    ARRAY['https://wellbeing.mohw.gov.tw/nor/mmap/'], ARRAY['doit'],
    NOW(), NOW(), 'two_d',
    'SELECT district AS x_axis, COUNT(*)::float AS data FROM public.heal_mental_health_sites WHERE county = ''臺北市'' GROUP BY district ORDER BY data DESC, district',
    NULL, 'taipei'
  ),
  (
    'mental_health_support_sites', NULL, ARRAY[9202],
    '{"mode":"byParam","byParam":{"xParam":"area_label"}}',
    'current', NULL,
    1, 'day', '衛生福利部',
    '顯示雙北各行政區心理健康支持機構數量。',
    '此元件統計臺北市與新北市各行政區心理健康支持機構數量，並可在地圖檢視機構名稱、類型、地址、電話、服務資訊與網站。',
    '可用於雙北心理健康支持資源配置與跨行政區比較，支援校園壓力與青壯世代心理支持議題。',
    ARRAY['https://wellbeing.mohw.gov.tw/nor/mmap/'], ARRAY['doit', 'ntpc'],
    NOW(), NOW(), 'two_d',
    'SELECT county || '' '' || district AS x_axis, COUNT(*)::float AS data FROM public.heal_mental_health_sites WHERE county IN (''臺北市'', ''新北市'') GROUP BY county, district ORDER BY data DESC, county, district',
    NULL, 'metrotaipei'
  );

UPDATE public.dashboards
SET components = array_append(components, 9201), updated_at = NOW()
WHERE index IN ('map-layers-taipei', 'map-layers-metrotaipei')
  AND NOT 9201 = ANY(components);

DO $$
DECLARE
  seq_name text;
BEGIN
  seq_name := pg_get_serial_sequence('public.components', 'id');
  IF seq_name IS NOT NULL THEN
    EXECUTE format('SELECT setval(%L, (SELECT MAX(id) FROM public.components), true)', seq_name);
  END IF;

  seq_name := pg_get_serial_sequence('public.component_maps', 'id');
  IF seq_name IS NOT NULL THEN
    EXECUTE format('SELECT setval(%L, (SELECT MAX(id) FROM public.component_maps), true)', seq_name);
  END IF;
END $$;

COMMIT;
