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


-- +goose down
DROP TABLE refresh_tokens;

