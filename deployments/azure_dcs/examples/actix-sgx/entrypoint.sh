#!/usr/bin/bash
set -eu
cat rust.sig | xxd -s 0x3c0 -l 32 -p -c 32
/restart_aesm.sh
gramine-sgx rust
