-- +goose Up
-- +goose StatementBegin
-- USERS
ALTER TABLE
    users
ALTER COLUMN
    id
SET
    DEFAULT uuid_generate_v4();

ALTER TABLE
    users
ALTER COLUMN
    created_at
SET
    DEFAULT NOW();

ALTER TABLE
    users
ALTER COLUMN
    updated_at
SET
    DEFAULT NOW();

-- FEEDS
ALTER TABLE
    feeds
ALTER COLUMN
    id
SET
    DEFAULT uuid_generate_v4();

ALTER TABLE
    feeds
ALTER COLUMN
    created_at
SET
    DEFAULT NOW();

ALTER TABLE
    feeds
ALTER COLUMN
    updated_at
SET
    DEFAULT NOW();

-- FEED_FOLLOWS
ALTER TABLE
    feed_follows
ALTER COLUMN
    id
SET
    DEFAULT uuid_generate_v4();

ALTER TABLE
    feed_follows
ALTER COLUMN
    created_at
SET
    DEFAULT NOW();

ALTER TABLE
    feed_follows
ALTER COLUMN
    updated_at
SET
    DEFAULT NOW();

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
-- USERS
ALTER TABLE
    users
ALTER COLUMN
    id DROP DEFAULT;

ALTER TABLE
    users
ALTER COLUMN
    created_at DROP DEFAULT;

ALTER TABLE
    users
ALTER COLUMN
    updated_at DROP DEFAULT;

-- FEEDS
ALTER TABLE
    feeds
ALTER COLUMN
    id DROP DEFAULT;

ALTER TABLE
    feeds
ALTER COLUMN
    created_at DROP DEFAULT;

ALTER TABLE
    feeds
ALTER COLUMN
    updated_at DROP DEFAULT;

-- FEED_FOLLOWS
ALTER TABLE
    feed_follows
ALTER COLUMN
    id DROP DEFAULT;

ALTER TABLE
    feed_follows
ALTER COLUMN
    created_at DROP DEFAULT;

ALTER TABLE
    feed_follows
ALTER COLUMN
    updated_at DROP DEFAULT;

-- +goose StatementEnd
