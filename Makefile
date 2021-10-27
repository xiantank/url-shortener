DATABASE_HOST ?= localhost
DATABASE_PORT ?= 3306
DATABASE_NAME ?= url_shortener
DATABASE_USER ?= root
DATABASE_PASSWORD ?= root
REDIS_PORT ?= 6379

init:
	make redis-bloom-up
	mysql -u root -e "create database ${DATABASE_NAME}" -p${DATABASE_PASSWORD}
	make db-up

run:
	go run ./main.go

db-clean:
	mysql -u root -e "drop database if exists ${DATABASE_NAME}" -p${DATABASE_PASSWORD}

db-up:
	migrate -path ./migrations -database "mysql://${DATABASE_USER}:${DATABASE_PASSWORD}@tcp(${DATABASE_HOST}:${DATABASE_PORT})/${DATABASE_NAME}?charset=utf8mb4" -verbose up

db-down:
	migrate -path ./migrations -database "mysql://${DATABASE_USER}:${DATABASE_PASSWORD}@tcp(${DATABASE_HOST}:${DATABASE_PORT})/${DATABASE_NAME}?charset=utf8mb4" -verbose down -all

redis-bloom-up:
	docker start redis-redisbloom 2>/dev/null 1>/dev/null || docker run -d -p ${REDIS_PORT}:6379 --name redis-redisbloom redislabs/rebloom:latest
#	docker kill redis-redisbloom

clean:
	make db-clean

reinit:
	make clean
	make init
