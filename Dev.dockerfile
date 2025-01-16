# syntax=docker/dockerfile:experimental
FROM golang:1.23-bullseye

LABEL maintainer="Phoops info@phoops.it"
LABEL project="Bellatrix"
LABEL environment="development"

RUN apt update && apt install -y build-essential

# Install air for hot reloadings
RUN go install github.com/air-verse/air@latest

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz

WORKDIR /app

ADD . /app

# Start hot reload
CMD make start-dev
