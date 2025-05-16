-- name: MarkFeed :exec
UPDATE feeds
SET last_fetched_at = $1
WHERE id = $2;