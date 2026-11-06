-- Initial seeder data matching GeoTrack frontend mock dataset
INSERT INTO entities (id, code, name, type, status, latitude, longitude, description, created_at, updated_at)
VALUES
    ('c1f7a091-8849-4bc2-9e96-a94f30ad8151', 'VH-001', 'Fleet Alpha - Truck 01', 'vehicle', 'active', -6.2088, 106.8456, 'Logistics delivery heavy vehicle serving West Jakarta sector', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
    ('d2a8b102-9950-4cd3-af07-b05e41be9262', 'IOT-101', 'Cold Storage Sensor 01', 'iot_device', 'active', -6.1754, 106.8272, 'Temperature and humidity telemetry device for frozen goods', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
    ('e3b9c213-0061-4de4-bf18-c16f52cf0373', 'FAC-201', 'Central Warehouse Jakarta', 'facility', 'active', -6.2297, 106.8073, 'Main distribution and inventory center Sudirman area', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
    ('f4ca1324-1172-4ef5-c029-d27a63d01484', 'INF-301', 'Base Transceiver Station North', 'infrastructure', 'maintenance', -6.1214, 106.7741, 'Telecommunication tower facility undergoing antenna calibration', NOW() - INTERVAL '7 days', NOW() - INTERVAL '1 hour'),
    ('a5db2435-2283-4fa6-d130-e38b74e12595', 'VH-002', 'Fleet Beta - Van 02', 'vehicle', 'inactive', -6.2615, 106.7811, 'Secondary light courier van in South Jakarta reserve depot', NOW() - INTERVAL '12 days', NOW() - INTERVAL '2 days'),
    ('b6ec3546-3394-4ab7-e241-f49c85f23606', 'IOT-102', 'Air Quality Monitor Station', 'iot_device', 'maintenance', -6.1950, 106.8230, 'Atmospheric particulate PM2.5 monitoring sensor module', NOW() - INTERVAL '4 days', NOW() - INTERVAL '3 hours'),
    ('c7fd4657-4405-4bc8-f352-a50d96a34717', 'FAC-202', 'Distribution Hub Bekasi', 'facility', 'active', -6.2383, 106.9756, 'East Jabodetabek cargo sorting and transit terminal', NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days'),
    ('d8ae5768-5516-4cd9-0463-b61ea7b45828', 'INF-302', 'Fiber Optic Junction Box', 'infrastructure', 'active', -6.2146, 106.8451, 'Core underground optical distribution junction point', NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days')
ON CONFLICT (code) DO NOTHING;

-- Seed matching activity logs
INSERT INTO activity_logs (id, entity_id, entity_code, entity_name, action, details, actor, created_at)
VALUES
    ('a9b2c3d4-e5f6-7890-abcd-ef1234567890', 'f4ca1324-1172-4ef5-c029-d27a63d01484', 'INF-301', 'Base Transceiver Station North', 'status_changed', 'Status changed from active to maintenance for antenna tuning', 'Admin Dispatcher', NOW() - INTERVAL '1 hour'),
    ('b8c3d4e5-f6a7-8901-bcde-f12345678901', 'b6ec3546-3394-4ab7-e241-f49c85f23606', 'IOT-102', 'Air Quality Monitor Station', 'status_changed', 'Scheduled maintenance initiated', 'System', NOW() - INTERVAL '3 hours'),
    ('c7d4e5f6-a7b8-9012-cdef-012345678902', 'c1f7a091-8849-4bc2-9e96-a94f30ad8151', 'VH-001', 'Fleet Alpha - Truck 01', 'created', 'Entity registered to fleet registry', 'System', NOW() - INTERVAL '3 days')
ON CONFLICT DO NOTHING;

