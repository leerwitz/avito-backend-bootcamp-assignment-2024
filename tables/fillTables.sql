INSERT INTO house (address, "year", developer, created_at, update_at) VALUES
('123 Elm Street, Springfield', 1999, 'Springfield Developers Inc.', '2023-08-01T10:00:00.000Z', '2023-08-01T10:00:00.000Z'),
('456 Maple Avenue, Shelbyville', 2005, 'Shelbyville Construction Co.', '2023-08-01T11:00:00.000Z', '2023-08-01T11:00:00.000Z'),
('789 Oak Boulevard, Capital City', 2010, 'Capital City Builders Ltd.', '2023-08-01T12:00:00.000Z', '2023-08-01T12:00:00.000Z'),
('101 Birch Lane, Ogdenville', 2015, 'Ogdenville Homes LLC', '2023-08-01T13:00:00.000Z', '2023-08-01T13:00:00.000Z');

INSERT INTO flat (house_id, price, rooms, flat_num, "status", moderator_id) VALUES
(1, 100000, 3, 101, 'created', NULL),
(1, 150000, 4, 102, 'approved', 1),
(2, 120000, 2, 201, 'on moderation', 2),
(2, 130000, 3, 202, 'declined', 2),
(3, 110000, 2, 301, 'created', NULL),
(3, 200000, 5, 302, 'approved', 3),
(4, 170000, 4, 401, 'on moderation', 4),
(4, 180000, 4, 402, 'declined', 4);