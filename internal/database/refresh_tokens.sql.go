// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: refresh_tokens.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createRefreshToken = `-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at) VALUES
(
  $1,
  NOW(),
  NOW(),
  $2,
  NOW() + INTERVAL '60 day'
) RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at
`

type CreateRefreshTokenParams struct {
	Token  string
	UserID uuid.UUID
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, createRefreshToken, arg.Token, arg.UserID)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}

const getUserByToken = `-- name: GetUserByToken :one
SELECT user_id
  from refresh_tokens 
 WHERE revoked_at is NULL
   AND token = $1
`

func (q *Queries) GetUserByToken(ctx context.Context, token string) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getUserByToken, token)
	var user_id uuid.UUID
	err := row.Scan(&user_id)
	return user_id, err
}

const revokeToken = `-- name: RevokeToken :one
UPDATE refresh_tokens
   SET revoked_at = NOW(),
       updated_at = NOW()
 WHERE token = $1
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at
`

func (q *Queries) RevokeToken(ctx context.Context, token string) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, revokeToken, token)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}
