#!/usr/bin/env bash

set -eEuxo pipefail

docker compose down
sudo rm -rf ./cache/* ./data/*
