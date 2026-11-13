-- name: DiagramCreate :one
INSERT INTO diagrams (id, title)
VALUES ($1, $2)
RETURNING *;

-- name: DiagramList :many
SELECT
    diagrams.*,
    diagram_collaborators.permission,
    CASE
        WHEN diagram_collaborators.permission = 'OWNER' AND EXISTS (
            SELECT 1
            FROM diagram_collaborators dc2
            WHERE dc2.diagram_id = diagrams.id
                AND dc2.permission != 'OWNER'
        ) THEN true
        ELSE false
    END AS has_collaborators
FROM diagram_collaborators
JOIN diagrams
    ON diagrams.id = diagram_collaborators.diagram_id
WHERE diagram_collaborators.user_id = $1
ORDER BY diagrams.title ASC;

-- name: DiagramGet :one
SELECT * FROM diagrams
WHERE id = $1;

-- name: DiagramGetByUser :one
SELECT diagrams.*, diagram_collaborators.permission
FROM diagrams
JOIN diagram_collaborators
    ON diagram_collaborators.diagram_id = diagrams.id
        AND diagram_collaborators.user_id = $1
WHERE diagrams.id = $2;

-- name: DiagramSetTitle :exec
UPDATE diagrams
SET title = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: DiagramDelete :exec
DELETE FROM diagrams
WHERE id = $1;
