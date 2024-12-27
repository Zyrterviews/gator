-- name: CreateUser :one
INSERT INTO
    users (name)
VALUES
    ($1)
RETURNING
    *;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    name = $1;

-- name: GetUsers :many
SELECT
    *
FROM
    users;

-- name: DeleteUsers :exec
DELETE FROM
    users;
