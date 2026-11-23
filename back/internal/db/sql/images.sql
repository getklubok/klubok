-- name: ImageCreate :one
INSERT INTO images (id, diagram_id, blob_hash, preview, original_name, mime)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ImageGet :one
SELECT * FROM images
WHERE id = $1
AND diagram_id = $2;

-- name: ImagePreviewListByDiagram :many
SELECT id, preview FROM images
WHERE diagram_id = $1;

-- name: ImageListByDiagram :many
SELECT * FROM images
WHERE diagram_id = $1;

-- name: ImageExistsByBlobHash :one
SELECT EXISTS (
    SELECT 1 FROM images
    WHERE blob_hash = $1
);

-- name: ImageDeleteForShape :one
DELETE FROM images
WHERE id = $1
AND diagram_id = $2
RETURNING blob_hash, mime;

-- name: ImageDeleteForDiagram :many
DELETE FROM images
WHERE diagram_id = $1
RETURNING blob_hash, mime;

-- name: ImageBulkInsert :copyfrom
INSERT INTO images (id, diagram_id, blob_hash, preview, original_name, mime, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);
