FROM golang:alpine

RUN apk update && apk add git

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app/src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/build/app ./cmd/

FROM scratch
WORKDIR /app

COPY --from=0 /app/build/app .

CMD ["/app/app", "-key", "$ETTERNA_API_KEY"]