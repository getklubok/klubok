-- name: AuthCodeUpsert :exec
INSERT INTO auth_codes (id, user_id, user_email, code)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id)
DO UPDATE SET
    code = EXCLUDED.code;

-- name: AuthCodeDelete :exec
DELETE FROM auth_codes
WHERE id = $1;

-- name: AuthCodeFind :one
SELECT * FROM auth_codes
WHERE user_email = $1
    AND code = $2
LIMIT 1;

-- name: AuthCodeCancelOld :exec
DELETE FROM auth_codes
WHERE created_at < $1;
