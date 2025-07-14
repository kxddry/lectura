-- Drop foreign key constraint first
ALTER TABLE transcribed DROP CONSTRAINT IF EXISTS fk_transcribed_uuid;


-- Drop table
DROP TABLE IF EXISTS transcribed;
