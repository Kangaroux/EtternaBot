#!/bin/sh
# Applies the latest database migrations. Must be ran from within a container.

echo "Applying migrations..."

for i in 1 2 3 4 5; do
    ./bin/migrate \
        -path migrations/ \
        -database postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$DATABASE_HOST/$POSTGRES_DB?sslmode=disable \
        up

    if [ $? -eq 0 ]; then
        echo "Migrations applied successfully."
        exit 0
    fi

    echo "Failed. Retrying ($i of 5)..."

    sleep 1
done

echo "Migrations could not be applied."
exit 1