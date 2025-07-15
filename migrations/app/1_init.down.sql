-- Drop foreign key constraint first
ALTER TABLE summarized DROP CONSTRAINT IF EXISTS fk_summarized_uuid;

-- Drop table
DROP TABLE IF EXISTS summarized;

-- Drop foreign key constraint first
ALTER TABLE transcribed DROP CONSTRAINT IF EXISTS fk_transcribed_uuid;

-- Drop table
DROP TABLE IF EXISTS transcribed;


-- Drop indexes
DROP INDEX IF EXISTS idx_files_user_id;
DROP INDEX IF EXISTS idx_files_status;

-- Drop table
DROP TABLE IF EXISTS files;
