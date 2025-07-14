-- Drop foreign key constraint first
ALTER TABLE summarized DROP CONSTRAINT IF EXISTS fk_summarized_uuid;

-- Drop table
DROP TABLE IF EXISTS summarized;
