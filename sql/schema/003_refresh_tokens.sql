-- +goose up
CREATE TABLE refresh_tokens (
  token TEXT PRIMARY KEY,
  user_id uuid NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  revoked_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

ALTER TABLE refresh_tokens 
ALTER COLUMN created_at SET NOT NULL,
ALTER COLUMN updated_at SET NOT NULL;

-- +goose down
DROP TABLE refresh_tokens;

