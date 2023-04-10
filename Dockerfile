# syntax=docker/dockerfile:1

# Alpine is chosen for its small footprint
# compared to Ubuntu
FROM golang:1.20.1-alpine

WORKDIR /usr/src/app

RUN go install github.com/cosmtrek/air@latest
RUN go get -u -t github.com/volatiletech/sqlboiler/v4@latest
RUN go get -u -t github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
COPY . .

EXPOSE 3000

RUN go mod tidy && go mod vendor