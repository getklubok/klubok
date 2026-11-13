-- +goose up
CREATE TABLE IF NOT EXISTS "user_sso_identities" (
    "id" CHAR(26) PRIMARY KEY,
    "user_id" CHAR(26) NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
    "issuer" TEXT NOT NULL,
    "sub" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    UNIQUE ("issuer", "sub")
);

CREATE INDEX IF NOT EXISTS "user_sso_identities__user_id__idx" ON "user_sso_identities" ("user_id");

-- +goose down
DROP TABLE IF EXISTS "user_sso_identities";
