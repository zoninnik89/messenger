INSERT INTO apps (id, name, secret)
VALUES (1,'test', 'test')
ON CONFLICT DO NOTHING;