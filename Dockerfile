# Build the app in its own container to cache dependencies
FROM golang:alpine
RUN apk update && apk add curl git
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /app/src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/build/app ./cmd/bot/

# Create a new container for running the app
FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=0 /app/build/app .
COPY --from=0 /app/src/migrate.sh .
COPY --from=0 /app/src/migrate_down.sh .
COPY --from=0 /app/src/run.sh .
COPY --from=0 /app/src/bin ./bin
COPY --from=0 /app/src/migrations ./migrations
ENTRYPOINT ["./run.sh"]