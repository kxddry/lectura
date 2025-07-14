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