#!/bin/bash

set -e

export GOOS=linux
export GOARCH=amd64

NAME=polyhedron
BUILD_DIR="build"

mkdir -p "$BUILD_DIR"

for app in server client; do
  echo "building $NAME $app"
  OUT_FILE="$BUILD_DIR/$NAME-$app-$GOOS-$GOARCH"
  go build -o "$OUT_FILE" "$ROOT_DIR/$app/"*
done

ansible-playbook -i ./hosts ./config/playbook.yaml