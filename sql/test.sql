INSERT INTO map_role_api (role_id, api_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'super_admin'),
    id
FROM apis
ON CONFLICT (role_id, api_id) DO NOTHING;
