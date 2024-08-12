CREATE TYPE flat_status AS ENUM ('created', 'approved', 'declined', 'on moderation');

CREATE TABLE IF NOT EXISTS house (
    id SERIAL PRIMARY KEY,
    address VARCHAR(1000) NOT NULL,
    "year" INTEGER NOT NULL CHECK ("year" >= 0),
    developer VARCHAR(1000),
    created_at DATE,
    update_at DATE
);

CREATE TABLE IF NOT EXISTS flat (
    id INTEGER NOT NULL CHECK (id >= 1),
    house_id INTEGER NOT NULL REFERENCES house(id),
    price INTEGER NOT NULL CHECK (price >= 0),
    rooms INTEGER NOT NULL CHECK (rooms >= 1),
    "status" flat_status,
    PRIMARY KEY (house_id, id) 
);