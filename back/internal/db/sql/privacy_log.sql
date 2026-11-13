-- name: PrivacyLogCreate :exec
INSERT INTO privacy_log (user_id, ip, privacy_version)
VALUES ($1, $2, $3);
