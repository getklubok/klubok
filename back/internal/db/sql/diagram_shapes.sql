-- name: DiagramShapeUpsert :exec
INSERT INTO diagram_shapes (id, diagram_id, data)
VALUES ($1, $2, $3)
ON CONFLICT (id, diagram_id)
DO UPDATE SET
    data = EXCLUDED.data,
    updated_at = NOW()
RETURNING *;

-- name: DiagramShapeBulkInsert :copyfrom
INSERT INTO diagram_shapes (id, diagram_id, data, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: DiagramShapeDelete :exec
DELETE FROM diagram_shapes
WHERE id = $1
AND diagram_id = $2;

-- name: DiagramShapeList :many
SELECT * FROM diagram_shapes
WHERE diagram_id = $1;
