-- name: UserCreate :one
INSERT INTO users (id, email, username, activate_code, is_fake)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UserUpdate :one
UPDATE users
SET email = $2,
    username = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UserFindByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: UserFindByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: UserFindByName :one
SELECT * FROM users
WHERE username = $1
LIMIT 1;

-- name: UserUpdateLastLoginAt :exec
UPDATE users SET
    last_login_at = NOW()
WHERE id = $1;

-- name: UserCancelActivateCode :exec
UPDATE users SET
    activate_code = NULL
WHERE id = $1;

-- name: UserDeleteNotActivated :exec
DELETE FROM users
WHERE created_at < $1
    AND activate_code IS NOT NULL;
