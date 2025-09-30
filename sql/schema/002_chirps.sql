-- +goose up
CREATE TABLE chirps (
id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
body TEXT NOT NULL,
user_id uuid NOT NULL,
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW(),
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose down
DROP TABLE chirps;
