#! /bin/sh

set -e

cd "`dirname "$0"`/src"
npm run make_css
npm run install_icons
npm run make_css
go build -o ../dist/piaf -tags release -ldflags="-extldflags=-Wl,-no_warn_duplicate_libraries" .

if [ "$1" == --install -o "$1" == -i ]; then
  echo Installing: may prompt for sudo password
  sudo install -m 0755 ../dist/piaf /usr/local/bin/
fi
