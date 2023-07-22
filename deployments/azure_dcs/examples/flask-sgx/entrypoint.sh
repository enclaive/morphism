#!/usr/bin/bash
set -eu

/restart_aesm.sh
cat app.sig | xxd -s 0x3c0 -l 32 -p -c 32
gramine-sgx app
