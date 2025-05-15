-- name: GetAllFeeds :many
SELECT feeds.name, feeds.url, users.name AS username
FROM feeds
LEFT JOIN users
ON feeds.user_id = users.id
GROUP BY feeds.name, feeds.url, username;