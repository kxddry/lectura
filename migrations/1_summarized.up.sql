-- Create `summarized` table
CREATE TABLE summarized (
                            id SERIAL PRIMARY KEY,
                            uuid TEXT NOT NULL UNIQUE,
                            text TEXT NOT NULL CHECK (octet_length(text) <= 1048576) -- max 1MB
);

ALTER TABLE summarized
    ADD CONSTRAINT fk_summarized_uuid FOREIGN KEY (uuid) REFERENCES files(uuid) ON DELETE CASCADE;