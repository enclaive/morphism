FROM golang:latest as pre-main

ENV CGO_ENABLED=0

COPY ./examples/premain-examples/ /app/
WORKDIR /app/

RUN mkdir /data && \
    go build -o pre-main -ldflags="-extldflags=-static -w" .


FROM rust:latest as build-env

WORKDIR /usr/src/test_serverless/
COPY ./examples/actix/test_serverless/ .

RUN cargo install --path .

FROM debian:buster-slim
RUN adduser --system --no-create-home nonroot
RUN apt-get update && apt-get install -y ca-certificates
	#&&
    #apt-get -y install openssl && 
    
COPY --from=pre-main \
    /app/pre-main \
    /app/
      
COPY --from=build-env  \
      /usr/local/cargo/bin/test_serverless /usr/local/bin/test_serverless

COPY ./examples/certs/cert.pem	/etc/ssl/certs/

USER nonroot     
WORKDIR /app/

CMD ["/app/pre-main"]
