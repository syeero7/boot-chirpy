-- +goose up
CREATE TABLE users (
id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
email TEXT UNIQUE NOT NULL,
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW()
);

ALTER TABLE users
ADD COLUMN hashed_password TEXT;

UPDATE users SET hashed_password = 'unset' 
WHERE hashed_password IS NULL;

ALTER TABLE users 
ALTER COLUMN hashed_password SET NOT NULL;

-- +goose down
DROP TABLE users;
