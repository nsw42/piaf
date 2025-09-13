#! /bin/sh

set -e

cd "`dirname "$0"`/src"
npm run make_css
npm run install_icons
npm run make_css
export CGO_CXXFLAGS="`pkg-config --cflags soundtouch`"
export CGO_LDFLAGS="`pkg-config --libs soundtouch`"
go build -o ../dist/piaf -tags release -ldflags="-extldflags=-Wl,-no_warn_duplicate_libraries" .
