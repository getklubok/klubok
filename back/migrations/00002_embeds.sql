-- +goose up
CREATE TABLE IF NOT EXISTS "diagram_embeds" (
    "id" CHAR(26) PRIMARY KEY,
    "diagram_id" CHAR(26) NOT NULL REFERENCES "diagrams" ("id") ON DELETE CASCADE,
    "user_id" CHAR(26) NOT NULL REFERENCES "users" ("id"),

    "token" CHAR(64) NOT NULL UNIQUE,
    "is_active" BOOLEAN NOT NULL DEFAULT true,
    "allowed_origins" TEXT[],
    "embed_options" JSONB NOT NULL DEFAULT '{}',

    "views_count" INTEGER NOT NULL DEFAULT 0,
    "last_viewed_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_diagram_embeds_token ON "diagram_embeds" ("token") WHERE "is_active" = true;
CREATE INDEX IF NOT EXISTS idx_diagram_embeds_diagram_id ON "diagram_embeds" ("diagram_id");

-- +goose down
DROP INDEX IF EXISTS idx_diagram_embeds_diagram_id;
DROP INDEX IF EXISTS idx_diagram_embeds_token;
DROP TABLE IF EXISTS "diagram_embeds";