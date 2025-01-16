# syntax=docker/dockerfile:experimental
FROM golang:1.23-bullseye as builder

WORKDIR /build
ADD . /build

RUN apt update && apt install -y build-essential
RUN --mount=type=ssh go mod download &&  make build-bellatrix

FROM alpine:3.17

LABEL maintainer="Phoops info@phoops.it"
LABEL project="Bellatrix"
LABEL environment="production"

WORKDIR /app
COPY --from=builder /build/bin/bellatrix /app
RUN ls -lah && chmod +x bellatrix && pwd
EXPOSE 8000

CMD ["./bellatrix", "sync"]
