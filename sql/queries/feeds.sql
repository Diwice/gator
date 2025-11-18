-- name: CreateFeed :one
INSERT INTO feeds(created_at, updated_at, name, url, user_id)
VALUES(
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING *;
-- name: GetFeeds :many
SELECT * FROM feeds;
-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE url = $1;
-- name: MarkFeedFetched :exec
UPDATE feeds
SET updated_at = $2, last_fetched_at = $2
WHERE feeds.id = $1;
-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;
