#!/bin/sh
# Applies the database migrations. Must be ran from within the `bot` docker container

./bin/migrate \
    -path migrations/ \
    -database postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@db/$POSTGRES_DB?sslmode=disable \
    up