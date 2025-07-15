CREATE TABLE files (
                       id SERIAL PRIMARY KEY,
                       uuid TEXT NOT NULL UNIQUE,
                       user_id INTEGER NOT NULL,
                       og_filename TEXT NOT NULL,
                       og_extension TEXT NOT NULL,
                       status SMALLINT NOT NULL CHECK (status IN (0, 1, 2)) -- uploaded, transcribed, summarized
);

CREATE INDEX idx_files_user_id ON files(user_id);
CREATE INDEX idx_files_status ON files(status);

-- Create `summarized` table
CREATE TABLE summarized (
                            id SERIAL PRIMARY KEY,
                            uuid TEXT NOT NULL UNIQUE,
                            text TEXT NOT NULL CHECK (octet_length(text) <= 1048576) -- max 1MB
);

ALTER TABLE summarized
    ADD CONSTRAINT fk_summarized_uuid FOREIGN KEY (uuid) REFERENCES files(uuid) ON DELETE CASCADE;



-- Create `transcribed` table
CREATE TABLE transcribed (
                             id SERIAL PRIMARY KEY,
                             uuid TEXT NOT NULL UNIQUE,
                             text TEXT NOT NULL CHECK (octet_length(text) <= 1048576), -- max 1MB
                             language TEXT NOT NULL
);

ALTER TABLE transcribed
    ADD CONSTRAINT fk_transcribed_uuid FOREIGN KEY (uuid) REFERENCES files(uuid) ON DELETE CASCADE;