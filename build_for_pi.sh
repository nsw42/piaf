#! /bin/bash

set -e

npm run make_css

PLATFORM=aarch64
cd $(dirname $0)
IMAGE_NAME=piaf-$PLATFORM
docker build --platform linux/$PLATFORM -f Dockerfile.pi -t $IMAGE_NAME .

container=$(docker create $IMAGE_NAME)
rm -rf dist/$PLATFORM
mkdir -p dist/$PLATFORM
docker cp $container:/go/src/dist/piaf ./dist/$PLATFORM/piaf
docker rm -v $container > /dev/null
