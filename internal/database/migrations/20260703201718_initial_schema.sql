-- +goose Up
-- +goose StatementBegin
CREATE TABLE "Users" (
    "id" BIGSERIAL PRIMARY KEY,
    "firstname" TEXT NOT NULL,
    "lastname" TEXT NOT NULL,
    "pronouns" TEXT NOT NULL DEFAULT '',
    "joinDate" TIMESTAMP WITH TIME ZONE NOT NULL,
    "setupComplete" BOOLEAN NOT NULL DEFAULT FALSE,
    "archived" BOOLEAN,
    "notes" TEXT NOT NULL DEFAULT '',
    "admin" BOOLEAN NOT NULL DEFAULT FALSE,
    "forceArchive" BOOLEAN,
    "cardTag" TEXT NOT NULL DEFAULT ''
);
-- +goose StatementEnd

-- +goose Down
SELECT 'down SQL query';
