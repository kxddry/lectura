
-- Create `transcribed` table
CREATE TABLE transcribed (
                             id SERIAL PRIMARY KEY,
                             uuid TEXT NOT NULL UNIQUE,
                             text TEXT NOT NULL CHECK (octet_length(text) <= 1048576), -- max 1MB
                             language TEXT NOT NULL
);

ALTER TABLE transcribed
    ADD CONSTRAINT fk_transcribed_uuid FOREIGN KEY (uuid) REFERENCES files(uuid) ON DELETE CASCADE;