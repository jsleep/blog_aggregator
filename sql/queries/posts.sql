-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, published_at, url, feed_id, title, description)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
WITH feed_follows AS (
    SELECT * FROM feed_follows
    WHERE feed_follows.user_id = $1
)
SELECT
    feeds.name AS feed_name,
    users.name AS user_name,
    posts.*
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
INNER JOIN users ON feed_follows.user_id = users.id
INNER JOIN posts ON feed_follows.feed_id = posts.feed_id
ORDER BY posts.published_at DESC
LIMIT $2;