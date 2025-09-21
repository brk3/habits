#!/usr/bin/env bash

docker run --rm -it \
  -v ./kanidm:/root/.config/kanidm:ro \
  -v ./cache:/root/.cache docker.io/kanidm/tools:1.7.3 \
  /sbin/kanidm $@
