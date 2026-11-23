-- +goose up
CREATE TABLE IF NOT EXISTS "images" (
    "id" CHAR(26) PRIMARY KEY,
    "diagram_id" CHAR(26) NOT NULL REFERENCES "diagrams" ("id") ON DELETE CASCADE,
    "blob_hash" CHAR(64) NOT NULL,
    "preview" BYTEA NOT NULL,
    "original_name" TEXT NOT NULL,
    "mime" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_images_diagram_id" ON "images" ("diagram_id");

CREATE INDEX IF NOT EXISTS "idx_images_blob_hash" ON "images" ("blob_hash");

-- +goose down
DROP TABLE IF EXISTS "images";
