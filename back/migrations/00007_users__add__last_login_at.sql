-- +goose up
ALTER TABLE "users" ADD COLUMN "last_login_at" TIMESTAMPTZ;

-- +goose down
ALTER TABLE "users" DROP COLUMN "last_login_at";
