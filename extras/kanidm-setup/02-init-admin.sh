#!/usr/bin/env bash

set -eEuxo pipefail

docker compose up -d

docker run --rm -it \
  -v $(pwd)/data:/data \
  -v $(pwd)/server.toml:/server.toml \
  docker.io/kanidm/server:latest \
  /sbin/kanidmd recover-account -c /server.toml admin

docker run --rm -it \
  -v $(pwd)/data:/data \
  -v $(pwd)/server.toml:/server.toml \
  docker.io/kanidm/server:latest \
  /sbin/kanidmd recover-account -c /server.toml idm_admin
