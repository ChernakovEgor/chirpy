-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at) VALUES
(
  $1,
  NOW(),
  NOW(),
  $2,
  NOW() + INTERVAL '60 day'
) RETURNING *;

-- name: GetUserByToken :one
SELECT user_id
  from refresh_tokens 
 WHERE revoked_at is NULL
   AND token = $1;

-- name: RevokeToken :one
UPDATE refresh_tokens
   SET revoked_at = NOW(),
       updated_at = NOW()
 WHERE token = $1
RETURNING *;
