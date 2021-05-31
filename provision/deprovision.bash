#!/bin/bash

multipass stop vm1 vm2 lb
multipass delete vm1 vm2 lb
multipass purge