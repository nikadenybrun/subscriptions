FROM golang:alpine

WORKDIR /app

RUN apk add --no-cache bash

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o subscriptions ./cmd

COPY config.env ./

ENV CONFIG_PATH=./config.env

COPY wait-for-it.sh ./
RUN chmod +x ./wait-for-it.sh

EXPOSE 8080

CMD ["./wait-for-it.sh", "postgres:5432", "--", "./subscriptions"]