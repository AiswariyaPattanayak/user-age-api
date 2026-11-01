-- name: CreateUser :one
INSERT INTO users (name, dob)
VALUES ($1, $2)
RETURNING id, name, dob;

-- name: GetUser :one
SELECT id, name, dob
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, dob
FROM users
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: UpdateUser :one
UPDATE users
SET name = $2, dob = $3
WHERE id = $1
RETURNING id, name, dob;

-- name: DeleteUser :execrows
DELETE FROM users WHERE id = $1;
