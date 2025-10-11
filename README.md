# habits

Habits is built on the idea that small, consistent actions yield the best results.

[![Test Status](https://github.com/brk3/habits/actions/workflows/test.yml/badge.svg)](https://github.com/brk3/habits/actions/workflows/test.yml)
[![Docker Build](https://github.com/brk3/habits/actions/workflows/docker-latest.yml/badge.svg)](https://github.com/brk3/habits/actions/workflows/docker-latest.yml)
[![Frontend Build](https://github.com/brk3/habits/actions/workflows/frontend-docker-latest.yml/badge.svg)](https://github.com/brk3/habits/actions/workflows/frontend-docker-latest.yml)

![Screenshot](./extras/screenshot.png)

## Install
```
curl -fsSL https://raw.githubusercontent.com/brk3/habits/main/install.sh | bash
```

## Development
```bash
# Build the project
make build

# Start a habits-server
make server

# Start a frontend
make frontend

# Track a habit with a note
habits track guitar "practiced riffs"

# > Access frontend on localhost:5173/habits/{habit}
```

## Configuring
See [config.yaml](./config.yaml)

## Hosting
A `docker-compose.yml` is included for a full working setup. It's fronted by a Caddy reverse proxy
which also hosts the `habits-frontend`.

This is secured by a wildcard letsencrypt cert, based on this [guide](https://blog.mni.li/posts/internal-tls-with-caddy).

```bash
# Fetch your wildcard cert
docker exec acme.sh --register-account -m my@example.com --server letsencrypt --set-default-ca
docker exec acme.sh --issue -k 4096 -d aiectomy.xyz -d '*.aiectomy.xyz' --dns dns_acmedns

# Bring up the habits stack
docker compose up -d
```