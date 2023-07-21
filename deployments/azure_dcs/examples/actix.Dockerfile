# Specifes name of the project and executable
ARG projectName

# build stage
FROM rust:latest AS builder
ARG projectName

COPY ./examples/actix-sgx/build .
COPY ./examples/actix-sgx/$projectName/ ./$projectName/
RUN apt-get update && xargs -a packages.txt -r apt-get install -y && rm -rf packages.txt /var/lib/apt/lists/*

# Compile executable in release mode
RUN cargo install --path ./$projectName/

# final stage
FROM gramineproject/gramine:v1.4
ARG projectName
# Specifes subdirectory in /entrypoint/ for web files, e.g. *.html, *.js, ...

COPY ./examples/actix-sgx/packages.txt .
RUN apt-get update && xargs -a packages.txt -r apt-get install -y && apt-get install -y --no-install-recommends libsgx-dcap-default-qpl && rm -rf packages.txt /var/lib/apt/lists/*
COPY ./examples/sgx_default_qcnl.conf /etc/
# Copy executable
COPY --from=builder /$projectName/target/release/$projectName /entrypoint/app
# Copy required files
COPY ./examples/actix-sgx/rust.manifest.template /entrypoint/
COPY ./examples/actix-sgx/entrypoint.sh /entrypoint/
COPY ./examples/premain /app/premain

WORKDIR /entrypoint/
RUN gramine-sgx-gen-private-key 
RUN gramine-manifest -Darch_libdir=/lib/x86_64-linux-gnu rust.manifest.template rust.manifest 
RUN gramine-sgx-sign --manifest rust.manifest --output rust.manifest.sgx
RUN gramine-sgx-get-token --output rust.token --sig rust.sig
EXPOSE 8080/tcp
ENTRYPOINT [ "/entrypoint/entrypoint.sh" ]
