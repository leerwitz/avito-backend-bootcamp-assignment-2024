DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'flat_status') THEN
        CREATE TYPE flat_status AS ENUM ('created', 'approved', 'declined', 'on moderation');
    END IF;
END $$;

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

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    user_type VARCHAR(50) NOT NULL CHECK (user_type IN ('client', 'moderator'))
);


DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_house_id' AND relkind = 'i') THEN
        CREATE INDEX idx_house_id ON flat (house_id);
    END IF;
END $$;

