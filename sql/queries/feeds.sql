-- name: CreateFeed :one
INSERT INTO
    feeds (name, url, user_id)
VALUES
    ($1, $2, $3)
RETURNING
    *;

-- name: GetFeed :one
SELECT
    feeds.id,
    feeds.created_at,
    feeds.updated_at,
    feeds.name,
    feeds.url,
    feeds.user_id,
    users.name AS user_name
FROM
    feeds
    JOIN users ON feeds.user_id = users.id
WHERE
    feeds.url = $1;

-- name: GetFeeds :many
SELECT
    feeds.id,
    feeds.created_at,
    feeds.updated_at,
    feeds.name,
    feeds.url,
    feeds.user_id,
    users.name AS user_name
FROM
    feeds
    JOIN users ON feeds.user_id = users.id;

-- name: MarkFeedFetched :exec
UPDATE
    feeds
SET
    last_fetched_at = NOW(),
    updated_at = NOW()
WHERE
    id = $1;

-- name: GetNextFeedToFetch :one
SELECT
    *
FROM
    feeds
ORDER BY
    last_fetched_at ASC NULLS FIRST
LIMIT
    1;
