run:
	go run ./cmd/server

test:
	go test ./... -v

sqlc:
	sqlc generate

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v
