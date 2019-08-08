.PHONY: migrations bot

default: db bot

bot:
	docker-compose up -d bot

db:
	docker-compose up -d db

build:
	docker-compose build bot

psql:
	docker-compose exec db psql

stop:
	docker-compose stop

testdb:
	docker-compose -f docker-compose-test.yml up db-test

# Starts the test database and runs the go tests using the test db. The test db is reset
# when this is ran. Use `make testdb` to start up the test db without destroying the
# volume to inspect data after a test
test:
	docker-compose -f docker-compose-test.yml up -V --build --exit-code-from bot-test

# Runs migrations on the production database
migrations:
	docker-compose run bot ./migrate.sh