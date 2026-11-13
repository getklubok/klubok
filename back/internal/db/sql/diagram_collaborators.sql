-- name: DiagramCollaboratorUpsert :one
INSERT INTO diagram_collaborators (id, diagram_id, user_id, permission)
VALUES ($1, $2, $3, $4)
ON CONFLICT (diagram_id, user_id)
DO UPDATE SET
    permission = EXCLUDED.permission,
    updated_at = NOW()
RETURNING *;

-- name: DiagramCollaboratorBulkInsert :copyfrom
INSERT INTO diagram_collaborators (id, diagram_id, user_id, permission, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DiagramCollaboratorList :many
SELECT diagram_collaborators.*, users.username AS user_name
FROM diagram_collaborators
JOIN users
    ON users.id = diagram_collaborators.user_id
WHERE diagram_collaborators.diagram_id = $1
    AND diagram_collaborators.permission != 'OWNER'
ORDER BY users.username ASC;

-- name: DiagramCollaboratorFlatList :many
SELECT * FROM diagram_collaborators
WHERE diagram_id = $1
ORDER BY created_at;

-- name: DiagramCollaboratorExist :one
SELECT * FROM diagram_collaborators
WHERE diagram_id = $1
    AND user_id = $2
LIMIT 1;

-- name: DiagramCollaboratorDelete :exec
DELETE FROM diagram_collaborators
WHERE diagram_id = $1
    AND id = $2;
