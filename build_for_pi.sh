#! /bin/bash

usage() {
  echo "$0 [PLATFORM] [CROSS_BUILDER_IMAGE]"
  exit 0
}

set -e

npm run make_css

if [ -z "$1" ]; then
  PLATFORM=aarch64
elif [ "$1" = "-h" -o "$1" = "--help" ]; then
  usage
else
  PLATFORM=$1
fi

if [ -z "$2" ]; then
  CROSS_BUILDER=go-cross-builder-1.24-alpine3.22-$PLATFORM
else
  CROSS_BUILDER=$2
fi

if [ -z "$3" ]; then
  TARGETPLATFORM=linux/$PLATFORM
else
  TARGETPLATFORM=$3
fi

cd $(dirname $0)
IMAGE_NAME=piaf-$PLATFORM
docker build --platform $TARGETPLATFORM --build-arg CROSS_BUILDER=$CROSS_BUILDER -f Dockerfile.pi -t $IMAGE_NAME .

container=$(docker create $IMAGE_NAME)
rm -rf dist/$PLATFORM
mkdir -p dist/$PLATFORM
docker cp $container:/go/src/dist/piaf ./dist/$PLATFORM/piaf
docker rm -v $container > /dev/null
