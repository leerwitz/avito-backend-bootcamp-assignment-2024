-- CREATE TYPE flat_status AS ENUM ('created', 'approved', 'declined', 'on moderation');

CREATE TABLE IF NOT EXISTS house (
    id SERIAL PRIMARY KEY,
    address VARCHAR(1000) NOT NULL,
    "year" INTEGER NOT NULL CHECK ("year" >= 0),
    developer VARCHAR(1000),
    created_at VARCHAR(255),
    update_at VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS flat (
    id SERIAL PRIMARY KEY,
    house_id INTEGER NOT NULL REFERENCES house(id),
    price INTEGER NOT NULL CHECK (price >= 0),
    rooms INTEGER NOT NULL CHECK (rooms >= 1),
    flat_num INTEGER NOT NULL CHECK (flat_num >= 1),
    "status" flat_status,
    moderator_id INTEGER,   
    CONSTRAINT unique_house_flat UNIQUE (house_id, flat_num)
);

-- CREATE INDEX idx_house_id ON flat (house_id);

