# syntax=docker/dockerfile:experimental
FROM golang:1.23-bullseye as builder

WORKDIR /build
ADD . /build

RUN apt update && apt install -y build-essential
RUN --mount=type=ssh go mod download &&  make build-token_retriever

FROM alpine:3.17

LABEL maintainer="Phoops info@phoops.it"
LABEL project="Wobcom Token Retriever"
LABEL environment="production"

WORKDIR /app
COPY --from=builder /build/bin/token_retriever /app
RUN ls -lah && chmod +x token_retriever && pwd
EXPOSE 8000

CMD ["./token_retriever"]
