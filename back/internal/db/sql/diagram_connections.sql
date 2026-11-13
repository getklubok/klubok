-- name: DiagramConnectionUpsert :exec
INSERT INTO diagram_connections (diagram_id, from_shape_id, to_shape_id)
VALUES ($1, $2, $3)
ON CONFLICT (diagram_id, from_shape_id, to_shape_id)
DO UPDATE SET
    from_shape_id = EXCLUDED.from_shape_id,
    to_shape_id = EXCLUDED.to_shape_id,
    updated_at = NOW()
RETURNING *;

-- name: DiagramConnectionBulkInsert :copyfrom
INSERT INTO diagram_connections (diagram_id, from_shape_id, to_shape_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: DiagramConnectionDelete :exec
DELETE FROM diagram_connections
WHERE diagram_id = $1
AND from_shape_id = $2
AND to_shape_id = $3;

-- name: DiagramConnectionList :many
SELECT * FROM diagram_connections
WHERE diagram_id = $1;
