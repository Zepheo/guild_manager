DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 from pg_type WHERE typname = 'raid_type') THEN
        CREATE TYPE raid_type AS ENUM ('mc','es','bwl');
    END IF;
END $$;

ALTER TABLE raids
ADD COLUMN IF NOT EXISTS raid raid_type NOT NULL DEFAULT 'mc';
