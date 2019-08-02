#!/bin/sh
# Reverts the last applied migration on the database.

echo "Reverting latest migration..."

for i in 1 2 3 4 5; do
    ./bin/migrate \
        -path migrations/ \
        -database postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$DATABASE_HOST/$POSTGRES_DB?sslmode=disable \
        down 1

    if [ $? -eq 0 ]; then
        echo "Migration reverted successfully."
        exit 0
    fi

    echo "Failed. Retrying ($i of 5)..."

    sleep 1
done

echo "Migration could not be reverted."
exit 1