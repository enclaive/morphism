#!/bin/bash
set -eu

/restart_aesm.sh
#gramine-sgx-get-token --output app.token --sig app.sig
gramine-sgx app
