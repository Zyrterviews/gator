-- name: CreatePost :one
INSERT INTO
    posts (title, url, description, published_at, feed_id)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetPosts :many
WITH feed_ids AS (
    SELECT
        id
    FROM
        feeds
    WHERE
        user_id = (
            SELECT
                users.id
            FROM
                users
            WHERE
                users.name = $1
        )
)
SELECT
    *
FROM
    posts
WHERE
    feed_id IN (
        SELECT
            id
        FROM
            feed_ids
    )
ORDER BY
    published_at DESC
LIMIT
    $2;
