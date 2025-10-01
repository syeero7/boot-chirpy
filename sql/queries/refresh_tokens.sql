-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ( $1, $2, $3);

-- name: GetUserByRefreshToken :one
SELECT u.* FROM refresh_tokens t
INNER JOIN users u ON u.id = t.user_id
WHERE t.token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = $2, updated_at = $3
WHERE token = $1;
