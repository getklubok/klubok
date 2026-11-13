-- name: UserSSOIdentityFind :one
SELECT * FROM user_sso_identities
WHERE issuer = $1 AND sub = $2
LIMIT 1;

-- name: UserSSOIdentityCreate :one
INSERT INTO user_sso_identities (id, user_id, issuer, sub)
VALUES ($1, $2, $3, $4)
RETURNING *;
