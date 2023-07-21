# change base image to enclave
# 
FROM gramineproject/gramine:v1.4

RUN apt-get update \
    && apt-get install -y  wget build-essential python3 

WORKDIR /app/

COPY ./examples/flask-sgx/app /app/
COPY ./examples/premain /app/premain
COPY ./examples/flask-sgx/app.manifest.template ./examples/flask-sgx/entrypoint.sh /app/
RUN apt install -y python3-pip && apt-get install -y --no-install-recommends libsgx-dcap-default-qpl
RUN pip3 install -r /app/requirements.txt
COPY ./examples/sgx_default_qcnl.conf /etc/
RUN gramine-sgx-gen-private-key &&\
    gramine-argv-serializer "/usr/bin/python3" "-m" "flask" "--app" "/app/app.py" "run" "--host=0.0.0.0" > args.txt &&\
    gramine-manifest -Darch_libdir=/lib/x86_64-linux-gnu app.manifest.template app.manifest &&\
    gramine-sgx-sign --manifest app.manifest --output app.manifest.sgx
RUN gramine-sgx-get-token --output app.token --sig app.sig
EXPOSE 5000/tcp

# ENTRYPOINT [ "/app/entrypoint.sh" ]
ENTRYPOINT [ "/bin/bash" ]
