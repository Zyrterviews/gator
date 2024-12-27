-- +goose Up
CREATE TABLE posts (
    id UUID NOT NULL DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    published_at TIMESTAMPTZ NOT NULL,
    feed_id UUID NOT NULL
);

-- +goose Down
DROP TABLE posts;
