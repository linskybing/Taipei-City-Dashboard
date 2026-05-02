BEGIN;

DELETE FROM public.dashboard_groups
WHERE dashboard_id IN (
  SELECT id FROM public.dashboards
  WHERE index IN ('illegal_parking_monitor_taipei', 'illegal_parking_monitor_metrotaipei')
);
DELETE FROM public.dashboards
WHERE index IN ('illegal_parking_monitor_taipei', 'illegal_parking_monitor_metrotaipei');
DELETE FROM public.query_charts
WHERE index IN ('illegal_parking_device_counts', 'illegal_parking_violation_counts');
DELETE FROM public.component_charts
WHERE index IN ('illegal_parking_device_counts', 'illegal_parking_violation_counts');
DELETE FROM public.components
WHERE index IN ('illegal_parking_device_counts', 'illegal_parking_violation_counts');
DELETE FROM public.component_maps
WHERE index IN ('illegal_parking_devices_taipei', 'illegal_parking_devices_metrotaipei');

INSERT INTO public.component_maps (id, index, title, type, source, size, icon, paint, property) VALUES
(9130, 'illegal_parking_devices_taipei', '臺北市違停設備點位', 'circle', 'geojson', NULL, NULL, '{"circle-color":"#2563eb","circle-radius":["interpolate",["linear"],["zoom"],9,4,11,6,13,8,15,10],"circle-stroke-color":"#ffffff","circle-stroke-width":1,"circle-opacity":0.88}', '[{"key":"city_label","name":"城市"},{"key":"district","name":"行政區"},{"key":"location_name","name":"設置位置"},{"key":"item","name":"取締項目"},{"key":"device_type","name":"設備類型"},{"key":"location_precision","name":"座標精度"},{"key":"data_time","name":"更新時間"}]'),
(9131, 'illegal_parking_devices_metrotaipei', '雙北違停設備點位', 'circle', 'geojson', NULL, NULL, '{"circle-color":["match",["get","city_label"],"臺北市","#2563eb","新北市","#16a34a","#6b7280"],"circle-radius":["interpolate",["linear"],["zoom"],9,4,11,6,13,8,15,10],"circle-stroke-color":"#ffffff","circle-stroke-width":1,"circle-opacity":0.88}', '[{"key":"city_label","name":"城市"},{"key":"district","name":"行政區"},{"key":"location_name","name":"設置位置"},{"key":"item","name":"取締項目"},{"key":"device_type","name":"設備類型"},{"key":"location_precision","name":"座標精度"},{"key":"data_time","name":"更新時間"}]');

INSERT INTO public.components (id, index, name) VALUES
(9030, 'illegal_parking_device_counts', '違規停車設備數量與位置'),
(9031, 'illegal_parking_violation_counts', '違規停車取締量統計');

INSERT INTO public.component_charts (index, color, types, unit) VALUES
('illegal_parking_device_counts', ARRAY['#2563eb','#16a34a'], ARRAY['BarChart','ColumnChart'], '處'),
('illegal_parking_violation_counts', ARRAY['#7c3aed'], ARRAY['BarChart','ColumnChart'], '件');

INSERT INTO public.query_charts (
  index, history_config, map_config_ids, map_filter, time_from, time_to,
  update_freq, update_freq_unit, source, short_desc, long_desc, use_case,
  links, contributors, created_at, updated_at, query_type, query_chart, query_history, city
) VALUES
('illegal_parking_device_counts', NULL, ARRAY[9130], '{"mode":"byParam","byParam":{"xParam":"city_label","yParam":null}}', 'current', NULL, 1, 'day', '違停設備資料', '顯示臺北市違規停車科技執法設備數量，並在地圖上顯示位置。', '以設備點位資料呈現臺北市違規停車科技執法設置數量與空間分布。', '先看設備總量，再直接對照地圖理解設備集中在哪些區域。', ARRAY['https://td.police.gov.taipei/cp.aspx?n=6FEDE1F9DBFD656E'], ARRAY['doit'], NOW(), NOW(), 'two_d', $$SELECT '臺北市' AS x_axis, COUNT(*)::int AS data FROM public.tran_illegal_parking_device_locations_metrotaipei WHERE city = 'taipei'$$, NULL, 'taipei'),
('illegal_parking_device_counts', NULL, ARRAY[9131], '{"mode":"byParam","byParam":{"xParam":"city_label","yParam":null}}', 'current', NULL, 1, 'day', '違停設備資料', '顯示雙北違規停車設備數量，並在地圖上顯示位置。', '整合臺北市科技執法與新北市自動偵測設備，對比雙北各自布點數量與位置。', '先比較臺北市與新北市設備量，再點圖表只看單一城市的設備分布。', ARRAY['https://td.police.gov.taipei/cp.aspx?n=6FEDE1F9DBFD656E','https://data.ntpc.gov.tw/datasets/bb59a616-3572-4e92-9d09-01bf422057a6'], ARRAY['doit','ntpc'], NOW(), NOW(), 'two_d', $$SELECT CASE city WHEN 'taipei' THEN '臺北市' WHEN 'new_taipei' THEN '新北市' ELSE city END AS x_axis, COUNT(*)::int AS data FROM public.tran_illegal_parking_device_locations_metrotaipei GROUP BY city ORDER BY array_position(ARRAY['臺北市','新北市'], CASE city WHEN 'taipei' THEN '臺北市' WHEN 'new_taipei' THEN '新北市' ELSE city END)$$, NULL, 'metrotaipei'),
('illegal_parking_violation_counts', NULL, NULL, NULL, 'current', NULL, 1, 'day', '違停統計資料', '顯示臺北市各行政區違規停車件數。', '以臺北市逐筆違停事件資料統計各行政區的違規停車舉發量。', '用來快速辨識臺北市哪些行政區違停舉發量較高。', ARRAY['https://data.taipei/dataset/detail?id=6df5ded8-ddfc-413b-93ea-4e26a5a77027'], ARRAY['doit'], NOW(), NOW(), 'two_d', $$SELECT district AS x_axis, COUNT(*)::int AS data FROM public.tran_illegal_parking_events_tpe GROUP BY district ORDER BY data DESC, x_axis$$, NULL, 'taipei'),
('illegal_parking_violation_counts', NULL, NULL, NULL, 'current', NULL, 1, 'day', '違停統計資料', '顯示臺北市各行政區違規停車件數，並附上新北市總計。', '整合臺北市逐筆違停事件與新北市月統計，讓使用者同時比較臺北市各行政區與新北市整體違停規模。', '可同時看出臺北市區域差異，以及新北市整體違停總量在雙北中的相對規模。', ARRAY['https://data.taipei/dataset/detail?id=6df5ded8-ddfc-413b-93ea-4e26a5a77027','https://data.ntpc.gov.tw/datasets/ecea9d2e-266d-4ade-b8dd-c88fc0eb9e87'], ARRAY['doit','ntpc'], NOW(), NOW(), 'two_d', $$WITH tpe AS (SELECT district AS x_axis, COUNT(*)::int AS data, 1 AS sort_order FROM public.tran_illegal_parking_events_tpe GROUP BY district), ntpc AS (SELECT '新北市總計'::text AS x_axis, SUM(quantity)::int AS data, 2 AS sort_order FROM public.tran_illegal_parking_stats_new_tpe) SELECT x_axis, data FROM (SELECT * FROM tpe UNION ALL SELECT * FROM ntpc) combined ORDER BY sort_order, data DESC, x_axis$$, NULL, 'metrotaipei');

INSERT INTO public.dashboards (id, index, name, components, icon, updated_at, created_at) VALUES
(9032, 'illegal_parking_monitor_taipei', '違規停車統計', '{9030,9031}', 'local_police', NOW(), NOW()),
(9033, 'illegal_parking_monitor_metrotaipei', '違規停車統計', '{9030,9031}', 'local_police', NOW(), NOW());

INSERT INTO public.dashboard_groups (dashboard_id, group_id) VALUES
(9032, 2),
(9033, 3);

COMMIT;
