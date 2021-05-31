#!/bin/bash

export GOOS=linux
export GOARCH=amd64

NAME=polyhedron
ROOT_DIR="/Users/maxedmands/Projects/polyhedron"
BUILD_DIR="$ROOT_DIR/build"
OUT_FILE="$BUILD_DIR/$NAME-$GOOS-$GOARCH"

mkdir -p "$BUILD_DIR"
go build -o "$OUT_FILE" "$ROOT_DIR/server.go"

ansible-playbook -i ./hosts ./config/playbook.yml