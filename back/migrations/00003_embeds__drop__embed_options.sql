-- +goose up
ALTER TABLE "diagram_embeds" DROP COLUMN "embed_options";

-- +goose down
ALTER TABLE "diagram_embeds" ADD COLUMN "embed_options" JSONB NOT NULL DEFAULT '{}';
