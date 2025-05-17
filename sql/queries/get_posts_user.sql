-- name: GetPostsByUser :many
SELECT *
FROM posts
INNER JOIN feed_follows
ON posts.feed_id = feed_follows.feed_id
WHERE feed_follows.feed_id IN (
    SELECT feed_id
    FROM feed_follows
    WHERE feed_follows.user_id = $1
)
ORDER BY published_at DESC
LIMIT $2;