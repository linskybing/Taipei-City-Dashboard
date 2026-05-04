BEGIN;

DELETE FROM public.dashboard_groups
WHERE dashboard_id IN (
  SELECT id FROM public.dashboards
  WHERE index IN ('traffic_construction_overview_taipei', 'traffic_construction_overview_metrotaipei')
);
DELETE FROM public.dashboards
WHERE index IN ('traffic_construction_overview_taipei', 'traffic_construction_overview_metrotaipei');
DELETE FROM public.query_charts
WHERE index IN ('traffic_construction_total', 'traffic_construction_by_district');
DELETE FROM public.component_charts
WHERE index IN ('traffic_construction_total', 'traffic_construction_by_district');
DELETE FROM public.components
WHERE index IN ('traffic_construction_total', 'traffic_construction_by_district');
DELETE FROM public.component_maps
WHERE index IN (
  'traffic_construction_district_taipei',
  'traffic_construction_sites_taipei',
  'traffic_construction_district_metrotaipei',
  'traffic_construction_sites_metrotaipei'
);

INSERT INTO public.component_maps (id, index, title, type, source, size, icon, paint, property) VALUES
(9140, 'traffic_construction_district_taipei', '行政區概覽', 'fill', 'geojson', NULL, NULL, '{"fill-color":"#2563eb","fill-opacity":0.26,"fill-outline-color":"#93c5fd"}', '[{"key":"city_label","name":"城市"},{"key":"district_display","name":"行政區"},{"key":"active_count","name":"施工數"}]'),
(9141, 'traffic_construction_sites_taipei', '施工點位', 'circle', 'geojson', NULL, NULL, '{"layer-default-visibility":"none","circle-color":"#2563eb","circle-radius":["interpolate",["linear"],["zoom"],9,3,11,5,13,7,15,9],"circle-stroke-color":"#ffffff","circle-stroke-width":1,"circle-opacity":0.88}', '[{"key":"city_label","name":"城市"},{"key":"district_display","name":"行政區"},{"key":"addr","name":"地址"},{"key":"npurp","name":"施工類型"},{"key":"app_name","name":"主管機關"},{"key":"tc_na","name":"施工單位"},{"key":"co_ti","name":"施工時段"},{"key":"start_date","name":"起始日期"},{"key":"end_date","name":"結束日期"},{"key":"status_label","name":"狀態"}]'),
(9142, 'traffic_construction_district_metrotaipei', '行政區概覽', 'fill', 'geojson', NULL, NULL, '{"fill-color":"#0f766e","fill-opacity":0.26,"fill-outline-color":"#5eead4"}', '[{"key":"city_label","name":"城市"},{"key":"district_display","name":"行政區"},{"key":"active_count","name":"施工數"}]'),
(9143, 'traffic_construction_sites_metrotaipei', '施工點位', 'circle', 'geojson', NULL, NULL, '{"layer-default-visibility":"none","circle-color":["match",["get","city_label"],"臺北市","#2563eb","新北市","#16a34a","#6b7280"],"circle-radius":["interpolate",["linear"],["zoom"],9,3,11,5,13,7,15,9],"circle-stroke-color":"#ffffff","circle-stroke-width":1,"circle-opacity":0.88}', '[{"key":"city_label","name":"城市"},{"key":"district_display","name":"行政區"},{"key":"addr","name":"地址"},{"key":"npurp","name":"施工類型"},{"key":"app_name","name":"主管機關"},{"key":"tc_na","name":"施工單位"},{"key":"co_ti","name":"施工時段"},{"key":"start_date","name":"起始日期"},{"key":"end_date","name":"結束日期"},{"key":"status_label","name":"狀態"}]');

INSERT INTO public.components (id, index, name) VALUES
(9040, 'traffic_construction_total', '雙北施工總量'),
(9041, 'traffic_construction_by_district', '行政區施工分布與位置');

INSERT INTO public.component_charts (index, color, types, unit) VALUES
('traffic_construction_total', ARRAY['#2563eb','#16a34a'], ARRAY['BarChart','ColumnChart'], '件'),
('traffic_construction_by_district', ARRAY['#0f766e'], ARRAY['BarChart','ColumnChart'], '件');

INSERT INTO public.query_charts (
  index, history_config, map_config_ids, map_filter, time_from, time_to,
  update_freq, update_freq_unit, source, short_desc, long_desc, use_case,
  links, contributors, created_at, updated_at, query_type, query_chart, query_history, city
) VALUES
('traffic_construction_total', NULL, NULL, NULL, 'current', NULL, 10, 'minute', '道路施工資料', '顯示臺北市目前進行中的施工總量。', '以臺北市道路施工 current table 統計今日仍在施工期間內的案件數，快速掌握整體施工規模。', '適合先看臺北市目前的施工總量，再切到雙北視角比較兩市規模差異。', ARRAY['https://data.taipei/dataset/detail?id=c208dabd-2da0-4e6d-8dbd-a004b9782b0a'], ARRAY['doit'], NOW(), NOW(), 'two_d', $$SELECT '臺北市' AS x_axis, COUNT(*)::int AS data FROM public.traffic_todayworks WHERE (cb_ad IS NULL OR cb_ad <= CURRENT_DATE) AND (cd_ad IS NULL OR cd_ad >= CURRENT_DATE)$$, NULL, 'taipei'),
('traffic_construction_total', NULL, NULL, NULL, 'current', NULL, 10, 'minute', '道路施工資料', '顯示雙北目前進行中的施工總量。', '整合臺北市與新北市施工 current table，僅統計今天仍在施工期間內的案件數，對比雙北即時施工量體。', '先比較雙北總施工量，再切到行政區 component 看施工熱區與實際位置。', ARRAY['https://data.taipei/dataset/detail?id=c208dabd-2da0-4e6d-8dbd-a004b9782b0a','https://data.ntpc.gov.tw/datasets/96b6101b-c033-4834-8bd5-e312651db7a0'], ARRAY['doit','ntpc'], NOW(), NOW(), 'two_d', $$WITH combined AS (SELECT '臺北市'::text AS city_label FROM public.traffic_todayworks WHERE (cb_ad IS NULL OR cb_ad <= CURRENT_DATE) AND (cd_ad IS NULL OR cd_ad >= CURRENT_DATE) UNION ALL SELECT '新北市'::text AS city_label FROM public.traffic_todayworks_ntpe WHERE (cb_ad IS NULL OR cb_ad <= CURRENT_DATE) AND (cd_ad IS NULL OR cd_ad >= CURRENT_DATE)) SELECT city_label AS x_axis, COUNT(*)::int AS data FROM combined GROUP BY city_label ORDER BY array_position(ARRAY['臺北市','新北市'], city_label)$$, NULL, 'metrotaipei'),
('traffic_construction_by_district', NULL, ARRAY[9140, 9141], '{"mode":"byParam","byParam":{"xParam":"district","yParam":null}}', 'current', NULL, 10, 'minute', '道路施工資料', '顯示臺北市各行政區目前進行中的施工數量，並可切換到施工點位。', '先以行政區 bar chart 看臺北市哪裡施工最密集，點選後地圖會從行政區概覽切換成該區施工點位，顯示地址與施工詳細資訊。', '適合從區域熱點一路縮到單一施工點，快速理解施工位置與主管機關。', ARRAY['https://data.taipei/dataset/detail?id=c208dabd-2da0-4e6d-8dbd-a004b9782b0a'], ARRAY['doit'], NOW(), NOW(), 'two_d', $$SELECT c_name AS x_axis, COUNT(*)::int AS data FROM public.traffic_todayworks WHERE (cb_ad IS NULL OR cb_ad <= CURRENT_DATE) AND (cd_ad IS NULL OR cd_ad >= CURRENT_DATE) GROUP BY c_name ORDER BY data DESC, x_axis$$, NULL, 'taipei'),
('traffic_construction_by_district', NULL, ARRAY[9142, 9143], '{"mode":"byParam","byParam":{"xParam":"district","yParam":null}}', 'current', NULL, 10, 'minute', '道路施工資料', '顯示雙北各行政區目前進行中的施工數量，並可切換到施工點位。', '整合臺北市與新北市 current table，先以行政區 bar chart 找出雙北施工熱區，再點擊圖表切到該行政區的施工點位與詳細資訊。', '適合先看雙北哪個行政區施工最集中，再在地圖上追到實際施工案件。', ARRAY['https://data.taipei/dataset/detail?id=c208dabd-2da0-4e6d-8dbd-a004b9782b0a','https://data.ntpc.gov.tw/datasets/96b6101b-c033-4834-8bd5-e312651db7a0'], ARRAY['doit','ntpc'], NOW(), NOW(), 'two_d', $$SELECT c_name AS x_axis, COUNT(*)::int AS data FROM (SELECT c_name FROM public.traffic_todayworks WHERE (cb_ad IS NULL OR cb_ad <= CURRENT_DATE) AND (cd_ad IS NULL OR cd_ad >= CURRENT_DATE) UNION ALL SELECT c_name FROM public.traffic_todayworks_ntpe WHERE (cb_ad IS NULL OR cb_ad <= CURRENT_DATE) AND (cd_ad IS NULL OR cd_ad >= CURRENT_DATE)) works GROUP BY c_name ORDER BY data DESC, x_axis$$, NULL, 'metrotaipei');

INSERT INTO public.dashboards (id, index, name, components, icon, updated_at, created_at) VALUES
(9042, 'traffic_construction_overview_taipei', '道路施工概覽', '{9040,9041}', 'construction', NOW(), NOW()),
(9043, 'traffic_construction_overview_metrotaipei', '道路施工概覽', '{9040,9041}', 'construction', NOW(), NOW());

INSERT INTO public.dashboard_groups (dashboard_id, group_id) VALUES
(9042, 2),
(9043, 3);

COMMIT;
