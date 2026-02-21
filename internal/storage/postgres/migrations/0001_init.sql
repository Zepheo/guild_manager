CREATE TABLE IF NOT EXISTS characters (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    class TEXT
);

CREATE TABLE IF NOT EXISTS raids (
    id TEXT PRIMARY KEY,
    raid_date TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS reserves (
    id SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id),
    raid_id TEXT REFERENCES raids(id),
    item_id INT REFERENCES items(id)
);

CREATE TABLE IF NOT EXISTS bonus_history (
    id SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id),
    old_bonus INT,
    new_bonus INT,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
