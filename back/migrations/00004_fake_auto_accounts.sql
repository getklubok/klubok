-- +goose up
ALTER TABLE "users" ADD COLUMN "is_fake" BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose down
ALTER TABLE "users" DROP COLUMN "is_fake";
