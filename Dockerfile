# syntax=docker/dockerfile:1

# Alpine is chosen for its small footprint
# compared to Ubuntu
FROM golang:1.20.1-alpine

WORKDIR /usr/src/app

RUN go install github.com/cosmtrek/air@latest

COPY . .

EXPOSE 3000

RUN go mod tidy && go mod vendor