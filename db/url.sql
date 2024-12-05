-- name: CreateURL :exec
INSERT INTO urls (
    original_url,
    short_code,
    expires_at,
    is_custom
) VALUES (
             ?, ?, ?, ?
         );

-- name: GetCreatedURL :one
SELECT * FROM urls
WHERE id = LAST_INSERT_ID();

-- name: GetURLByShortCode :one
SELECT * FROM urls
WHERE short_code = ? AND expires_at > NOW()
LIMIT 1;

-- name: DeleteExpiredURLs :exec
DELETE FROM urls
WHERE expires_at <= NOW();

-- name: IsShortCodeAvailable :one
SELECT NOT EXISTS (
    SELECT 1 FROM urls
    WHERE short_code = ?
) AS is_available;
