redis:
	docker-compose -f ./docker/docker-compose.yml up -d --build redis

rm-redis:
	docker-compose -f ./docker/docker-compose.yml down

clean: rm-redis
	docker rmi -f redis