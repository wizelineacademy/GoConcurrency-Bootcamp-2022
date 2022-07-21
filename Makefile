build:
	mkdir -p functions
	GOOS=linux GOARCH=amd64 go build -o functions/main ./main.go

redis:
	docker-compose -f ./docker/docker-compose.yml up -d --build redis

rm-redis:
	docker-compose -f ./docker/docker-compose.yml down

clean: rm-redis
	docker rmi -f redis