FROM golang:latest as pre-main

ENV CGO_ENABLED=0

COPY ../../premain-vault/ /app/
WORKDIR /app/

RUN mkdir /data && \
    go build -o pre-main -ldflags="-extldflags=-static -w" .


FROM rust:latest as build-env

WORKDIR /usr/src/test_serverless/
COPY ./test_serverless .

RUN cargo install --path .

FROM debian:buster-slim

RUN apt-get update
	#&&
    #apt-get -y install openssl
    
COPY --from=pre-main \
    /app/pre-main \
    /app/
      
COPY --from=build-env  \
      /usr/local/cargo/bin/test_serverless /usr/local/bin/test_serverless
      
WORKDIR /app/

CMD ["/app/pre-main"]
