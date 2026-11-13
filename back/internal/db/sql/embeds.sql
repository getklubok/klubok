-- name: EmbedCreate :one
INSERT INTO diagram_embeds (id, diagram_id, user_id, token)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: EmbedGetByToken :one
SELECT * FROM diagram_embeds
WHERE token = $1;

-- name: EmbedGetByID :one
SELECT * FROM diagram_embeds
WHERE id = $1;

-- name: EmbedListByDiagram :many
SELECT * FROM diagram_embeds
WHERE diagram_id = $1
ORDER BY created_at DESC;

-- name: EmbedSoftDelete :exec
UPDATE diagram_embeds SET
    is_active = false,
    updated_at = NOW()
WHERE id = $1;

-- name: EmbedIncrementViews :exec
UPDATE diagram_embeds SET
    views_count = views_count + 1,
    last_viewed_at = NOW()
WHERE id = $1;
