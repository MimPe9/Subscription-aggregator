run:
    go run cmd/server/main.go

build:
    docker-compose build --no-cache

docker-up:
    docker-compose up

docker-down:
    docker-compose down
