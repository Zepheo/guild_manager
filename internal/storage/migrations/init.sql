-- Core character tracking
CREATE TABLE characters (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL, -- The character name from logs/csv
    class TEXT,
    current_bonus INT DEFAULT 0
);

-- Item catalog (to map names to IDs)
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    wowhead_id INT
);

-- Tracks the actual raids processed
CREATE TABLE raids (
    id SERIAL PRIMARY KEY,
    turtlogs_id TEXT UNIQUE NOT NULL, -- Extracted from the URL
    raid_date TIMESTAMP NOT NULL,
    processed_at TIMESTAMP DEFAULT NOW()
);

-- The "Intent": What people reserved
CREATE TABLE reserves (
    id SERIAL PRIMARY KEY,
    raid_id INT REFERENCES raids(id),
    character_id INT REFERENCES characters(id),
    item_id INT REFERENCES items(id)
);

-- The "Reality": Who actually got what
CREATE TABLE loot_drops (
    id SERIAL PRIMARY KEY,
    raid_id INT REFERENCES raids(id),
    item_id INT REFERENCES items(id),
    winner_id INT REFERENCES characters(id)
);

-- History of bonus changes for manual overrides and audit
CREATE TABLE bonus_history (
    id SERIAL PRIMARY KEY,
    character_id INT REFERENCES characters(id),
    change_amount INT,
    reason TEXT, -- e.g., "Raid [ID] missed roll", "Manual adjustment by Admin"
    created_at TIMESTAMP DEFAULT NOW()
);