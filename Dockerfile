FROM golang:1.25.0-alpine3.22 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./bin/api ./cmd/server/main.go

FROM alpine:latest

ENV ENVIRONMENT=compose

WORKDIR /app

COPY --from=build /app/bin/api .
COPY ./swagger ./swagger
COPY ./configs  ./configs

CMD [ "./api" ]
