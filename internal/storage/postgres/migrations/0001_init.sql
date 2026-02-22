CREATE TABLE IF NOT EXISTS characters (
    id INT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    class TEXT
);

CREATE TABLE IF NOT EXISTS raids (
    id INT PRIMARY KEY,
    raid_date TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS reserves (
    id SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id),
    raid_id INT REFERENCES raids(id),
    item_id INT REFERENCES items(id)
);

CREATE TABLE IF NOT EXISTS attendance (
    raid_id INT REFERENCES raids(id),
    character_id INT REFERENCES characters(id),
    PRIMARY KEY (raid_id, character_id)
);

CREATE TABLE IF NOT EXISTS loot_history (
    id SERIAL PRIMARY KEY,
    raid_id INT REFERENCES raids(id),
    item_id INT REFERENCES items(id),
    winner_id INT REFERENCES characters(id), -- Can be NULL if it was rot/disenchanted
    is_reserve_win BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS manual_bonus_adjustments (
    id SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id),
    amount INT,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
