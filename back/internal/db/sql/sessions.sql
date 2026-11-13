-- name: SessionCreate :one
INSERT INTO sessions (id, cid, uid, checked_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: SessionUpdateCheckedAt :one
UPDATE sessions
SET checked_at = $2
WHERE id = $1
RETURNING *;

-- name: SessionDeleteByID :exec
DELETE FROM sessions
WHERE id = $1;

-- name: SessionCancelOld :exec
DELETE FROM sessions
WHERE checked_at < $1;

-- name: SessionFindByUID :many
SELECT * FROM sessions
WHERE uid = $1
ORDER BY checked_at DESC;

-- name: SessionFindByCID :one
SELECT * FROM sessions
WHERE cid = $1
LIMIT 1;

-- name: SessionFindByID :one
SELECT * FROM sessions
WHERE id = $1
LIMIT 1;
