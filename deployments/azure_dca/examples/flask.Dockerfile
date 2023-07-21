# syntax=docker/dockerfile:1
FROM golang:latest as pre-main

ENV CGO_ENABLED=0

COPY ./examples/premain-examples/ /app/
WORKDIR /app/

RUN mkdir /data && \
    go build -o pre-main -ldflags="-extldflags=-static -w" .

FROM python:3.11.3-slim-buster
RUN adduser --system --no-create-home nonroot
RUN apt-get update && apt-get install -y ca-certificates
    
COPY --from=pre-main \
    /app/pre-main \
    /app/
      
WORKDIR /app/
COPY ./examples/flask/requirements.txt requirements.txt
COPY ./examples/certs/cert.pem	/etc/ssl/certs/
RUN pip3 install -r requirements.txt
COPY ./examples/flask/ .
USER nonroot 
CMD ["/app/pre-main","-m" , "flask", "run", "--host=0.0.0.0"]
