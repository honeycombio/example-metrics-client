#!/bin/bash

ROOT_DIR="/Users/maxedmands/Projects/polyhedron"
INFRA_DIR="$ROOT_DIR/provision"

for host in vm1 vm2 lb; do
  multipass launch -n=$host --cloud-init="$INFRA_DIR/cloud-init.yaml"
done