.PHONY: migrations

default: db bot

build:
	docker-compose build

stop:
	docker-compose stop

bot:
	docker-compose up -d bot

db:
	docker-compose up -d db

migrations:
	docker-compose run bot ./migrate.sh