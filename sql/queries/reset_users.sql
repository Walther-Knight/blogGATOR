-- name: ResetUsers :exec
DELETE FROM feed_follows;
DELETE FROM users;
DELETE FROM feeds;