# Specifes name of the project and executable
ARG projectName
# build stage
FROM golang:latest as build-env
ENV CGO_ENABLED=0

COPY ./examples/vault-sgx-plugin/ /app/
WORKDIR /app/
RUN go build -o premain  -buildmode=pie -ldflags="-extldflags=-static -w" ./cmd/premain-app
FROM rust:latest AS builder
ARG projectName

COPY ./examples/actix-sgx/build .
COPY ./examples/actix-sgx/test_serverless/ ./test_serverless/
RUN apt-get update && xargs -a packages.txt -r apt-get install -y && rm -rf packages.txt /var/lib/apt/lists/*

# Compile executable in release mode
RUN cargo install --path ./test_serverless/

# final stage
FROM gramineproject/gramine:v1.4
# Specifes subdirectory in /entrypoint/ for web files, e.g. *.html, *.js, ...

COPY ./examples/actix-sgx/packages.txt .
RUN apt-get update && xargs -a packages.txt -r apt-get install -y && apt-get install -y --no-install-recommends libsgx-dcap-default-qpl && rm -rf packages.txt /var/lib/apt/lists/*
COPY ./examples/sgx_default_qcnl.conf /etc/
# Copy executable

WORKDIR /app/
COPY --from=build-env /app/premain /app/
COPY --from=builder /test_serverless/target/release/test_serverless ./app

# Copy required files

COPY ./examples/actix-sgx/rust.manifest.template ./
COPY ./examples/actix-sgx/entrypoint.sh ./

RUN gramine-sgx-gen-private-key \
    && gramine-manifest -Darch_libdir=/lib/x86_64-linux-gnu rust.manifest.template rust.manifest \
    && gramine-sgx-sign --manifest rust.manifest --output rust.manifest.sgx
EXPOSE 8080/tcp
ENTRYPOINT [ "./entrypoint.sh" ]
