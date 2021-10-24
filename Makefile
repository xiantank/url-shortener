DATABASE_HOST ?= localhost
DATABASE_PORT ?= 3306
DATABASE_NAME ?= url_shortener
DATABASE_USER ?= root
DATABASE_PASSWORD ?= root

init:
	mysql -u root -e "create database ${DATABASE_NAME}" -p${DATABASE_PASSWORD}
	make db-up

clean:
	mysql -u root -e "drop database if exists ${DATABASE_NAME}" -p${DATABASE_PASSWORD}

db-up:
	migrate -path ./migrations -database "mysql://${DATABASE_USER}:${DATABASE_PASSWORD}@tcp(${DATABASE_HOST}:${DATABASE_PORT})/${DATABASE_NAME}?charset=utf8mb4" -verbose up

db-down:
	migrate -path ./migrations -database "mysql://${DATABASE_USER}:${DATABASE_PASSWORD}@tcp(${DATABASE_HOST}:${DATABASE_PORT})/${DATABASE_NAME}?charset=utf8mb4" -verbose down -all

reinit:
	make clean
	make init
