#!/bin/bash

set -e

for name in $(cat services); do
  docker build . -t ghcr.io/kxddry/lectura-$name -f $name/Dockerfile && docker push ghcr.io/kxddry/lectura-$name;
done
