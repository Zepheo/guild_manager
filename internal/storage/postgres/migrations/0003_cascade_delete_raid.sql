ALTER TABLE public.reserves DROP CONSTRAINT IF EXISTS reserves_raid_id_fkey;
ALTER TABLE public.reserves
ADD CONSTRAINT reserves_raid_id_fkey
    FOREIGN KEY (raid_id) REFERENCES raids(id) ON DELETE CASCADE;

ALTER TABLE public.attendance DROP CONSTRAINT IF EXISTS attendance_raid_id_fkey;
ALTER TABLE public.attendance
ADD CONSTRAINT attendance_raid_id_fkey
    FOREIGN KEY (raid_id) REFERENCES raids(id) ON DELETE CASCADE;

ALTER TABLE public.loot_history DROP CONSTRAINT IF EXISTS loot_history_raid_id_fkey;
ALTER TABLE public.loot_history
ADD CONSTRAINT loot_history_raid_id_fkey
    FOREIGN KEY (raid_id) REFERENCES raids(id) ON DELETE CASCADE;
