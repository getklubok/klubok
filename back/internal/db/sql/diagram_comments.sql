-- name: DiagramCommentCreate :one
INSERT INTO diagram_comments (id, diagram_id, diagram_shape_id, user_id, text)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DiagramCommentBulkInsert :copyfrom
INSERT INTO diagram_comments (id, diagram_id, diagram_shape_id, user_id, text, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DiagramCommentList :many
SELECT diagram_comments.*, users.username AS user_name
FROM diagram_comments
JOIN users
    ON users.id = diagram_comments.user_id
WHERE diagram_comments.diagram_id = $1
    AND diagram_comments.diagram_shape_id = $2
ORDER BY diagram_comments.created_at ASC;

-- name: DiagramCommentFlatList :many
SELECT * FROM diagram_comments
WHERE diagram_id = $1
ORDER BY created_at;

-- name: DiagramCommentClear :exec
DELETE FROM diagram_comments
WHERE diagram_id = $1
    AND diagram_shape_id = $2;

-- name: DiagramCommentStats :many
SELECT
    diagram_shape_id AS shape_id,
    COUNT(*) AS comment_count
FROM diagram_comments
WHERE diagram_id = $1
GROUP BY diagram_shape_id;
