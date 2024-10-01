-- name: Insert :one
INSERT INTO guest (id, message, created_at, updated_at, ip)
VALUES ($1, $2, $3, $3, $4)
RETURNING *;

-- name: FindAll :many
SELECT *
FROM guest
ORDER BY created_at DESC
LIMIT $1;

-- name: Count :one
SELECT COUNT(*) FROM guest;
