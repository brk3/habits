# habits

[![Build Status](https://github.com/brk3/habits/actions/workflows/test.yml/badge.svg)](https://github.com/brk3/habits/actions/workflows/test.yml)
![Release](https://github.com/brk3/habits/actions/workflows/release.yml/badge.svg)

A simple command-line tool for tracking habits, built in Go using [Cobra](https://github.com/spf13/cobra).

## Features

- Track a habit from the command line
- Outputs JSON for easy parsing or logging
- Includes tests and CI via GitHub Actions

## Getting Started

### Prerequisites

- Go 1.22 or higher
- `make` (optional)

### Clone and Build

```bash
git clone https://github.com/brk3/habits.git
cd habits
make build
