-- +goose up
ALTER TABLE "diagram_embeds" DROP COLUMN "allowed_origins";

-- +goose down
ALTER TABLE "diagram_embeds" ADD COLUMN "allowed_origins" TEXT[];
