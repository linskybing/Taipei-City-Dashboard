BEGIN;

DELETE FROM public.dashboard_groups
WHERE dashboard_id IN (
  SELECT id FROM public.dashboards
  WHERE id IN (9010, 9011)
     OR index IN ('tourism_parking_planning_taipei', 'tourism_parking_planning_metrotaipei')
);
DELETE FROM public.dashboards
WHERE id IN (9010, 9011)
   OR index IN ('tourism_parking_planning_taipei', 'tourism_parking_planning_metrotaipei');
DELETE FROM public.query_charts WHERE index IN (
  'tourism_parking_total_by_district',
  'tourism_parking_available_by_district',
  'tourism_parking_supply_balance_by_district',
  'tourism_parking_status_distribution',
  'tourism_parking_map'
);
DELETE FROM public.component_charts WHERE index IN (
  'tourism_parking_total_by_district',
  'tourism_parking_available_by_district',
  'tourism_parking_supply_balance_by_district',
  'tourism_parking_status_distribution',
  'tourism_parking_map'
);
DELETE FROM public.components
WHERE id IN (9010, 9011, 9012, 9013)
   OR index IN (
    'tourism_parking_total_by_district',
    'tourism_parking_available_by_district',
    'tourism_parking_supply_balance_by_district',
    'tourism_parking_status_distribution',
    'tourism_parking_map'
  );
DELETE FROM public.component_maps
WHERE id IN (9110, 9111, 9112, 9113)
   OR index IN (
    'tourism_parking_district_taipei',
    'tourism_onstreet_parking_taipei',
    'tourism_parking_district_metrotaipei',
    'tourism_onstreet_parking_metrotaipei'
  );

INSERT INTO public.component_maps (id, index, title, type, source, size, icon, paint, property) VALUES
(9110, 'tourism_parking_district_taipei', '行政區概覽', 'fill', 'geojson', NULL, NULL, '{"fill-color":"#2563eb","fill-opacity":0.28,"fill-outline-color":"#93c5fd"}', '[{"key":"district","name":"行政區"},{"key":"city","name":"城市"}]'),
(9111, 'tourism_onstreet_parking_taipei', '路邊車格點位', 'circle', 'geojson', NULL, NULL, '{"layer-default-visibility":"none","circle-color":["match",["get","status_label"],"空位","#16a34a","已停","#dc2626","狀態未知","#6b7280","不可用","#9ca3af","#6b7280"],"circle-radius":["interpolate",["linear"],["zoom"],9,2,11,4,13,6,15,8],"circle-stroke-color":"#ffffff","circle-stroke-width":1,"circle-opacity":0.86}', '[{"key":"name","name":"車格"},{"key":"district","name":"行政區"},{"key":"address","name":"路段"},{"key":"status_label","name":"狀態"},{"key":"data_time","name":"更新時間"}]'),
(9112, 'tourism_parking_district_metrotaipei', '行政區概覽', 'fill', 'geojson', NULL, NULL, '{"fill-color":"#0f766e","fill-opacity":0.28,"fill-outline-color":"#5eead4"}', '[{"key":"district","name":"行政區"},{"key":"city","name":"城市"}]'),
(9113, 'tourism_onstreet_parking_metrotaipei', '路邊車格點位', 'circle', 'geojson', NULL, NULL, '{"layer-default-visibility":"none","circle-color":["match",["get","status_label"],"空位","#16a34a","已停","#dc2626","狀態未知","#6b7280","不可用","#9ca3af","#6b7280"],"circle-radius":["interpolate",["linear"],["zoom"],9,2,11,4,13,6,15,8],"circle-stroke-color":"#ffffff","circle-stroke-width":1,"circle-opacity":0.86}', '[{"key":"name","name":"車格"},{"key":"city","name":"城市"},{"key":"district","name":"行政區"},{"key":"address","name":"路段"},{"key":"status_label","name":"狀態"},{"key":"data_time","name":"更新時間"}]');

INSERT INTO public.components (id, index, name) VALUES
(9010, 'tourism_parking_total_by_district', '行政區路邊車格總數'),
(9011, 'tourism_parking_supply_balance_by_district', '行政區空位與已占車格'),
(9012, 'tourism_parking_status_distribution', '路邊車格狀態分布'),
(9013, 'tourism_parking_map', '景點周邊路邊停車地圖');

INSERT INTO public.component_charts (index, color, types, unit) VALUES
('tourism_parking_total_by_district', ARRAY['#2563eb'], ARRAY['BarChart','ColumnChart'], '格'),
('tourism_parking_supply_balance_by_district', ARRAY['#16a34a','#dc2626'], ARRAY['ColumnChart'], '格'),
('tourism_parking_status_distribution', ARRAY['#16a34a','#dc2626','#6b7280','#9ca3af'], ARRAY['DonutChart','BarChart'], '格'),
('tourism_parking_map', ARRAY['#16a34a','#dc2626','#6b7280','#9ca3af'], ARRAY['MapLegend'], NULL);

INSERT INTO public.query_charts (
  index, history_config, map_config_ids, map_filter, time_from, time_to,
  update_freq, update_freq_unit, source, short_desc, long_desc, use_case,
  links, contributors, created_at, updated_at, query_type, query_chart, query_history, city
) VALUES
('tourism_parking_total_by_district', NULL, ARRAY[9110, 9111], '{"mode":"byParam","byParam":{"xParam":"district","yParam":null}}', 'current', NULL, 10, 'minute', '交通資料', '顯示臺北市各行政區的路邊車格總數。', '以臺北市路邊單一車格資料聚合各行政區供給量，協助駕駛在接近景點前先看哪一區整體停車供給較多。', '先用總量概覽找出候選行政區，再點圖表切到該區車格點位查看即時占用狀態。', ARRAY['https://tdx.transportdata.tw/'], ARRAY['doit'], NOW(), NOW(), 'two_d', $$SELECT district AS x_axis, COUNT(*)::int AS data FROM public.tran_onstreet_parking_realtime_tpe GROUP BY district ORDER BY data DESC, x_axis$$, NULL, 'taipei'),
('tourism_parking_total_by_district', NULL, ARRAY[9112, 9113], '{"mode":"byParam","byParam":{"xParam":"district","yParam":null}}', 'current', NULL, 10, 'minute', '交通資料', '顯示雙北各行政區的路邊車格總數。', '整合雙北路邊單一車格資料聚合各行政區供給量，協助駕駛在接近景點前先看哪一區整體停車供給較多。', '先用總量概覽找出候選行政區，再點圖表切到該區車格點位查看即時占用狀態。', ARRAY['https://tdx.transportdata.tw/','https://data.ntpc.gov.tw/datasets/54a507c4-c038-41b5-bf60-bbecb9d052c6'], ARRAY['doit','ntpc'], NOW(), NOW(), 'two_d', $$SELECT district AS x_axis, COUNT(*)::int AS data FROM (SELECT district FROM public.tran_onstreet_parking_realtime_tpe UNION ALL SELECT district FROM public.tran_onstreet_parking_realtime_new_tpe) parking GROUP BY district ORDER BY data DESC, x_axis$$, NULL, 'metrotaipei'),
('tourism_parking_supply_balance_by_district', NULL, ARRAY[9110, 9111], '{"mode":"byParam","byParam":{"xParam":"district","yParam":null}}', 'current', NULL, 10, 'minute', '交通資料', '顯示臺北市各行政區空位與已占車格數量。', '把臺北市路邊車格依行政區拆成空位與已占兩個系列，讓駕駛快速比較哪一區仍有可用供給。', '適合搭配地圖一起使用，點選任一行政區後只顯示該區所有 onstreet 車格點位。', ARRAY['https://tdx.transportdata.tw/'], ARRAY['doit'], NOW(), NOW(), 'three_d', $$SELECT district AS x_axis, status_group AS y_axis, COUNT(*)::int AS data FROM (SELECT district, CASE WHEN available_car = 1 THEN '空位' WHEN occupied_car = 1 THEN '已占' ELSE '其他' END AS status_group FROM public.tran_onstreet_parking_realtime_tpe) parking WHERE status_group IN ('空位','已占') GROUP BY district, status_group ORDER BY x_axis, array_position(ARRAY['空位','已占'], status_group)$$, NULL, 'taipei'),
('tourism_parking_supply_balance_by_district', NULL, ARRAY[9112, 9113], '{"mode":"byParam","byParam":{"xParam":"district","yParam":null}}', 'current', NULL, 10, 'minute', '交通資料', '顯示雙北各行政區空位與已占車格數量。', '把雙北路邊車格依行政區拆成空位與已占兩個系列，讓駕駛快速比較哪一區仍有可用供給。', '適合搭配地圖一起使用，點選任一行政區後只顯示該區所有 onstreet 車格點位。', ARRAY['https://tdx.transportdata.tw/','https://data.ntpc.gov.tw/datasets/54a507c4-c038-41b5-bf60-bbecb9d052c6'], ARRAY['doit','ntpc'], NOW(), NOW(), 'three_d', $$SELECT district AS x_axis, status_group AS y_axis, COUNT(*)::int AS data FROM (SELECT district, CASE WHEN available_car = 1 THEN '空位' WHEN occupied_car = 1 THEN '已占' ELSE '其他' END AS status_group FROM public.tran_onstreet_parking_realtime_tpe UNION ALL SELECT district, CASE WHEN available_car = 1 THEN '空位' WHEN occupied_car = 1 THEN '已占' ELSE '其他' END AS status_group FROM public.tran_onstreet_parking_realtime_new_tpe) parking WHERE status_group IN ('空位','已占') GROUP BY district, status_group ORDER BY x_axis, array_position(ARRAY['空位','已占'], status_group)$$, NULL, 'metrotaipei'),
('tourism_parking_status_distribution', NULL, ARRAY[9111], '{"mode":"byParam","byParam":{"xParam":"status_label","yParam":null}}', 'current', NULL, 10, 'minute', '交通資料', '顯示臺北市路邊車格目前是空位、已停、狀態未知或不可用。', '以路邊單一車格即時狀態分布呈現停車壓力，幫助駕駛理解熱門區域是否值得繼續前往。', '可快速聚焦某一種車格狀態，對照地圖查看這批點位的空間分布。', ARRAY['https://tdx.transportdata.tw/'], ARRAY['doit'], NOW(), NOW(), 'two_d', $$SELECT status_label AS x_axis, COUNT(*)::int AS data FROM public.tran_onstreet_parking_realtime_tpe GROUP BY status_label ORDER BY array_position(ARRAY['空位','已停','狀態未知','不可用'], status_label)$$, NULL, 'taipei'),
('tourism_parking_status_distribution', NULL, ARRAY[9113], '{"mode":"byParam","byParam":{"xParam":"status_label","yParam":null}}', 'current', NULL, 10, 'minute', '交通資料', '顯示雙北路邊車格目前是空位、已停、狀態未知或不可用。', '以路邊單一車格即時狀態分布呈現停車壓力，幫助駕駛理解熱門區域是否值得繼續前往。', '可快速聚焦某一種車格狀態，對照地圖查看這批點位的空間分布。', ARRAY['https://tdx.transportdata.tw/','https://data.ntpc.gov.tw/datasets/54a507c4-c038-41b5-bf60-bbecb9d052c6'], ARRAY['doit','ntpc'], NOW(), NOW(), 'two_d', $$SELECT status_label AS x_axis, COUNT(*)::int AS data FROM (SELECT status_label FROM public.tran_onstreet_parking_realtime_tpe UNION ALL SELECT status_label FROM public.tran_onstreet_parking_realtime_new_tpe) parking GROUP BY status_label ORDER BY array_position(ARRAY['空位','已停','狀態未知','不可用'], status_label)$$, NULL, 'metrotaipei'),
('tourism_parking_map', NULL, ARRAY[9110, 9111], NULL, 'current', NULL, 10, 'minute', '交通資料', '未點擊圖表時先顯示臺北市行政區概覽，點擊後切到該區路邊車格點位。', '地圖預設先用行政區面圖提供空間概覽，點擊圖表後再顯示指定行政區 onstreet 車格點位，並依即時占用狀態著色。', '適合從行政區層級快速縮到單一車格層級，支援旅遊前的停車決策。', ARRAY['https://tdx.transportdata.tw/'], ARRAY['doit'], NOW(), NOW(), 'map_legend', $$SELECT * FROM (VALUES ('空位','circle',''),('已停','circle',''),('狀態未知','circle',''),('不可用','circle','')) AS legend(name, type, icon)$$, NULL, 'taipei'),
('tourism_parking_map', NULL, ARRAY[9112, 9113], NULL, 'current', NULL, 10, 'minute', '交通資料', '未點擊圖表時先顯示雙北行政區概覽，點擊後切到該區路邊車格點位。', '地圖預設先用雙北行政區面圖提供空間概覽，點擊圖表後再顯示指定行政區 onstreet 車格點位，並依即時占用狀態著色。', '適合從行政區層級快速縮到單一車格層級，支援旅遊前的停車決策。', ARRAY['https://tdx.transportdata.tw/','https://data.ntpc.gov.tw/datasets/54a507c4-c038-41b5-bf60-bbecb9d052c6'], ARRAY['doit','ntpc'], NOW(), NOW(), 'map_legend', $$SELECT * FROM (VALUES ('空位','circle',''),('已停','circle',''),('狀態未知','circle',''),('不可用','circle','')) AS legend(name, type, icon)$$, NULL, 'metrotaipei');

INSERT INTO public.dashboards (id, index, name, components, icon, updated_at, created_at) VALUES
(9010, 'tourism_parking_planning_taipei', '觀光停車決策', '{9010,9011,9012,9013}', 'local_parking', NOW(), NOW()),
(9011, 'tourism_parking_planning_metrotaipei', '觀光停車決策', '{9010,9011,9012,9013}', 'local_parking', NOW(), NOW());

INSERT INTO public.dashboard_groups (dashboard_id, group_id) VALUES
(9010, 2),
(9011, 3);

COMMIT;
