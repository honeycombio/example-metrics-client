#!/bin/bash

for host in vm1 vm2 lb; do
  multipass launch -n=$host --cloud-init="provision/cloud-init.yaml"
done