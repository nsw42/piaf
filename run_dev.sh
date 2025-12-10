#! /bin/sh

set -e

ARGS=("$@")
N=$(( ${#ARGS[@]} - 1))

if [ "$N" -gt 0 ]; then
  for i in `seq 0 $N`; do
    if [ "${ARGS[$i]}" == "-d" ]; then
      j=$(( i+1 ))
      ARGS[$j]=$(abspath "${ARGS[$j]}")
    fi
  done
fi

cd "`dirname "$0"`/src"
export CGO_CXXFLAGS="`pkg-config --cflags soundtouch`"
export CGO_LDFLAGS="`pkg-config --libs soundtouch`"
go run -ldflags="-extldflags=-Wl,-no_warn_duplicate_libraries" . "${ARGS[@]}"
