-- name: ResetFeedFollows :exec
DELETE FROM feed_follows;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: ResetFeeds :exec
DELETE FROM feeds;