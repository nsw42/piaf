#! /bin/sh

set -e

cd "`dirname "$0"`/src"  # NB This means that -d DIR may not work as expected
export CGO_CXXFLAGS="`pkg-config --cflags soundtouch`"
export CGO_LDFLAGS="`pkg-config --libs soundtouch`"
go run -ldflags="-extldflags=-Wl,-no_warn_duplicate_libraries" . "$@"

