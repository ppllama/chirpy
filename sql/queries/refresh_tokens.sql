-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, revoked_at, user_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    NOW() + INTERVAL '60 days',
    NULL,
    $2
)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1
AND expires_at > NOW()
AND revoked_at IS NULL;

-- name: UpdateRevoke :exec
UPDATE refresh_tokens
SET updated_at = NOW(), revoked_at = NOW()
WHERE token = $1;