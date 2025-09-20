#!/usr/bin/env bash

set -eEuxo pipefail

docker run --rm -i -t \
  -v $(pwd)/data:/data \
  -v $(pwd)/server.toml:/server.toml \
  docker.io/kanidm/server:latest \
  kanidmd cert-generate -c /server.toml
