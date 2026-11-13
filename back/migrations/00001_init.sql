-- +goose up
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE
    IF NOT EXISTS "users" (
        "id" CHAR(26) PRIMARY KEY,
        "email" CITEXT NOT NULL UNIQUE,
        "username" CITEXT NOT NULL UNIQUE,
        "activate_code" CHAR(40),
        "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE TABLE
    IF NOT EXISTS "privacy_log" (
        "id" BIGSERIAL NOT NULL PRIMARY KEY,
        "user_id" CHAR(26) NOT NULL,
        "ip" TEXT NOT NULL,
        "privacy_version" TEXT NOT NULL,
        "action_at" TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE TABLE
    IF NOT EXISTS "auth_codes" (
        "id" CHAR(26) NOT NULL PRIMARY KEY,
        "user_id" CHAR(26) NOT NULL UNIQUE REFERENCES "users" ("id") ON DELETE CASCADE,
        "user_email" CITEXT NOT NULL,
        "code" CHAR(6) NOT NULL UNIQUE,
        "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE TABLE
    IF NOT EXISTS "sessions" (
        "id" CHAR(26) PRIMARY KEY,
        "cid" CHAR(30) NOT NULL UNIQUE,
        "uid" CHAR(26) NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
        "checked_at" TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS "sessions__uid__idx" ON "sessions" ("uid");

CREATE TABLE
    IF NOT EXISTS "diagrams" (
        "id" CHAR(26) PRIMARY KEY,
        "title" TEXT NOT NULL,
        "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE TABLE
    IF NOT EXISTS "diagram_collaborators" (
        "id" CHAR(26) PRIMARY KEY,
        "diagram_id" CHAR(26) NOT NULL REFERENCES "diagrams" ("id") ON DELETE CASCADE,
        "user_id" CHAR(26) NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
        "permission" TEXT NOT NULL CHECK ("permission" IN ('OWNER', 'VIEW', 'EDIT')),
        "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        UNIQUE ("diagram_id", "user_id")
    );

CREATE TABLE
    IF NOT EXISTS "diagram_shapes" (
        "id" CHAR(26) PRIMARY KEY,
        "diagram_id" CHAR(26) NOT NULL REFERENCES "diagrams" ("id") ON DELETE CASCADE,
        "data" JSONB NOT NULL,
        "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        UNIQUE ("id", "diagram_id")
    );

CREATE TABLE
    IF NOT EXISTS "diagram_connections" (
        "id" BIGSERIAL NOT NULL PRIMARY KEY,
        "diagram_id" CHAR(26) NOT NULL REFERENCES "diagrams" ("id") ON DELETE CASCADE,
        "from_shape_id" CHAR(26) NOT NULL,
        "to_shape_id" CHAR(26) NOT NULL,
        "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        UNIQUE ("diagram_id", "from_shape_id", "to_shape_id")
    );

CREATE TABLE
    IF NOT EXISTS "diagram_comments" (
        "id" CHAR(26) PRIMARY KEY,
        "diagram_id" CHAR(26) NOT NULL REFERENCES "diagrams" ("id") ON DELETE CASCADE,
        "diagram_shape_id" CHAR(26) NOT NULL REFERENCES "diagram_shapes" ("id") ON DELETE CASCADE,
        "user_id" CHAR(26) NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
        "text" TEXT NOT NULL,
        "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

-- +goose down
DROP TABLE IF EXISTS "diagram_comments";

DROP TABLE IF EXISTS "diagram_connections";

DROP TABLE IF EXISTS "diagram_shapes";

DROP TABLE IF EXISTS "diagram_collaborators";

DROP TABLE IF EXISTS "diagrams";

DROP TABLE IF EXISTS "sessions";

DROP TABLE IF EXISTS "auth_codes";

DROP TABLE IF EXISTS "privacy_log";

DROP TABLE IF EXISTS "users";
