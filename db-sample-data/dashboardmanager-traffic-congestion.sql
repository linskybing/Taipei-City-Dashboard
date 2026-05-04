BEGIN;

DELETE FROM public.dashboard_groups
WHERE dashboard_id IN (
  SELECT id FROM public.dashboards
  WHERE index IN ('traffic_congestion_overview_taipei', 'traffic_congestion_overview_metrotaipei')
);
DELETE FROM public.dashboards
WHERE index IN ('traffic_congestion_overview_taipei', 'traffic_congestion_overview_metrotaipei');
DELETE FROM public.query_charts
WHERE index IN ('traffic_congestion_ratio', 'traffic_congestion_segments');
DELETE FROM public.component_charts
WHERE index IN ('traffic_congestion_ratio', 'traffic_congestion_segments');
DELETE FROM public.components
WHERE index IN ('traffic_congestion_ratio', 'traffic_congestion_segments');
DELETE FROM public.component_maps
WHERE index IN ('traffic_congestion_lines_taipei', 'traffic_congestion_lines_metrotaipei');

INSERT INTO public.component_maps (id, index, title, type, source, size, icon, paint, property) VALUES
(9150, 'traffic_congestion_lines_taipei', '壅塞路段', 'line', 'geojson', NULL, NULL, '{"line-color":["match",["get","status_label"],"壅塞","#dc2626","車多","#f59e0b","順暢","#16a34a","#9ca3af"],"line-width":["interpolate",["linear"],["zoom"],9,2,11,3,13,4,15,6],"line-opacity":0.92}', '[{"key":"city_label","name":"城市"},{"key":"section_name","name":"路段名稱"},{"key":"source_type","name":"資料來源"},{"key":"avg_speed","name":"平均速度(km/h)"},{"key":"sample_count","name":"樣本數"},{"key":"status_label","name":"壅塞等級"},{"key":"data_time","name":"更新時間"}]'),
(9151, 'traffic_congestion_lines_metrotaipei', '雙北壅塞路段', 'line', 'geojson', NULL, NULL, '{"line-color":["match",["get","status_label"],"壅塞","#dc2626","車多","#f59e0b","順暢","#16a34a","#9ca3af"],"line-width":["interpolate",["linear"],["zoom"],9,2,11,3,13,4,15,6],"line-opacity":0.92}', '[{"key":"city_label","name":"城市"},{"key":"section_name","name":"路段名稱"},{"key":"source_type","name":"資料來源"},{"key":"avg_speed","name":"平均速度(km/h)"},{"key":"sample_count","name":"樣本數"},{"key":"status_label","name":"壅塞等級"},{"key":"data_time","name":"更新時間"}]');

INSERT INTO public.components (id, index, name) VALUES
(9050, 'traffic_congestion_ratio', '道路壅塞比例'),
(9051, 'traffic_congestion_segments', '道路壅塞路段分布與位置');

INSERT INTO public.component_charts (index, color, types, unit) VALUES
('traffic_congestion_ratio', ARRAY['#dc2626','#d1d5db'], ARRAY['BarPercentChart','GuageChart','IconPercentChart'], '段'),
('traffic_congestion_segments', ARRAY['#dc2626','#f59e0b','#16a34a','#9ca3af'], ARRAY['ColumnChart','BarChart','DonutChart'], '段');

INSERT INTO public.query_charts (
  index, history_config, map_config_ids, map_filter, time_from, time_to,
  update_freq, update_freq_unit, source, short_desc, long_desc, use_case,
  links, contributors, created_at, updated_at, query_type, query_chart, query_history, city
) VALUES
('traffic_congestion_ratio', NULL, NULL, NULL, 'current', NULL, 10, 'minute', '道路壅塞資料', '顯示臺北市目前壅塞路段與非壅塞路段比例。', '以臺北市官方 XML.GZ VD 即時速率資料換算道路壅塞等級，將目前路段分成壅塞與非壅塞兩類，快速掌握整體交通壓力。', '適合先用比例圖快速判斷臺北市當前交通是否緊繃，再切到路段分布組件查看實際壅塞位置。', ARRAY['https://tcgbusfs.blob.core.windows.net/blobtisv/GetVD.xml.gz'], ARRAY['doit'], NOW(), NOW(), 'percent', $$SELECT '道路壅塞比例' AS x_axis, y_axis, COUNT(*)::int AS data FROM (SELECT CASE WHEN status_label = '壅塞' THEN '壅塞路段' ELSE '非壅塞路段' END AS y_axis FROM public.traffic_road_congestion_tpe WHERE status_label IN ('壅塞','車多','順暢')) ratio GROUP BY y_axis ORDER BY array_position(ARRAY['壅塞路段','非壅塞路段'], y_axis)$$, NULL, 'taipei'),
('traffic_congestion_ratio', NULL, NULL, NULL, 'current', NULL, 10, 'minute', '道路壅塞資料', '顯示雙北目前壅塞路段與非壅塞路段比例。', '整合臺北市官方 VD 即時速率與新北市公車速度推估結果，將雙北目前路段分成壅塞與非壅塞兩類，快速掌握都會區交通壓力。', '適合先看雙北整體壅塞比例，再切到路段分布組件查出實際壅塞線段。', ARRAY['https://tcgbusfs.blob.core.windows.net/blobtisv/GetVD.xml.gz','https://tdx.transportdata.tw/api/basic/v2/Bus/RealTimeByFrequency/City/NewTaipei?%24format=JSON'], ARRAY['doit','ntpc'], NOW(), NOW(), 'percent', $$SELECT '道路壅塞比例' AS x_axis, y_axis, COUNT(*)::int AS data FROM (SELECT CASE WHEN status_label = '壅塞' THEN '壅塞路段' ELSE '非壅塞路段' END AS y_axis FROM (SELECT status_label FROM public.traffic_road_congestion_tpe UNION ALL SELECT status_label FROM public.traffic_road_congestion_ntpe) congestion WHERE status_label IN ('壅塞','車多','順暢')) ratio GROUP BY y_axis ORDER BY array_position(ARRAY['壅塞路段','非壅塞路段'], y_axis)$$, NULL, 'metrotaipei'),
('traffic_congestion_segments', NULL, ARRAY[9150], '{"mode":"byParam","byParam":{"xParam":"status_label","yParam":null}}', 'current', NULL, 10, 'minute', '道路壅塞資料', '顯示臺北市目前各壅塞等級的路段數量，並在地圖上以顏色標示實際路段。', '以臺北市官方 XML.GZ VD 即時速率資料建立實際路段線圖，圖表可切換壅塞、車多、順暢等級，地圖同步只顯示對應顏色路段。', '適合從整體分級數量一路縮到單一路段，快速找出壅塞熱點。', ARRAY['https://tcgbusfs.blob.core.windows.net/blobtisv/GetVD.xml.gz'], ARRAY['doit'], NOW(), NOW(), 'two_d', $$SELECT status_label AS x_axis, COUNT(*)::int AS data FROM public.traffic_road_congestion_tpe GROUP BY status_label ORDER BY array_position(ARRAY['壅塞','車多','順暢','資料不足'], status_label)$$, NULL, 'taipei'),
('traffic_congestion_segments', NULL, ARRAY[9151], '{"mode":"byParam","byParam":{"xParam":"status_label","yParam":null}}', 'current', NULL, 10, 'minute', '道路壅塞資料', '顯示雙北目前各壅塞等級的路段數量，並在地圖上以顏色標示實際路段。', '整合臺北市官方 VD 即時速率與新北市公車速度推估結果建立雙北路段線圖，圖表可切換壅塞、車多、順暢等級，地圖同步只顯示對應顏色路段。', '適合先看雙北壅塞等級分布，再在地圖上追到實際路段位置。', ARRAY['https://tcgbusfs.blob.core.windows.net/blobtisv/GetVD.xml.gz','https://tdx.transportdata.tw/api/basic/v2/Bus/RealTimeByFrequency/City/NewTaipei?%24format=JSON','https://tdx.transportdata.tw/api/basic/v2/Bus/Shape/City/NewTaipei?%24format=JSON'], ARRAY['doit','ntpc'], NOW(), NOW(), 'two_d', $$SELECT status_label AS x_axis, COUNT(*)::int AS data FROM (SELECT status_label FROM public.traffic_road_congestion_tpe UNION ALL SELECT status_label FROM public.traffic_road_congestion_ntpe) congestion GROUP BY status_label ORDER BY array_position(ARRAY['壅塞','車多','順暢','資料不足'], status_label)$$, NULL, 'metrotaipei');

INSERT INTO public.dashboards (id, index, name, components, icon, updated_at, created_at) VALUES
(9052, 'traffic_congestion_overview_taipei', '道路壅塞概覽', '{9050,9051}', 'traffic', NOW(), NOW()),
(9053, 'traffic_congestion_overview_metrotaipei', '道路壅塞概覽', '{9050,9051}', 'traffic', NOW(), NOW());

INSERT INTO public.dashboard_groups (dashboard_id, group_id) VALUES
(9052, 2),
(9053, 3);

COMMIT;
