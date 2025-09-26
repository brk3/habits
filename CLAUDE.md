# CLAUDE.md

This is a Go-based habits tracking application with a TypeScript frontend.

## Style Guidelines
- Use comments sparingly, the code should be self documenting
- Adhere to standard Go styling and conventions as enforced by 'make fmt'

## Build & Development Commands

### Backend (Go)
```bash
# Build the application
make build

# Start development server
make server

# Run tests with coverage
make test

# Format code
make fmt

# Clean build artifacts
make clean
```

### Frontend (TypeScript/Vite)
```bash
# Start frontend development server
make frontend
# OR
cd frontend && npm run dev
```

## Testing
- Go tests: `go test -cover ./...` (via `make test`)
- Test files follow `*_test.go` pattern

## Configuration
- Main config: `config.yaml` (OIDC, nudges, server settings)
- Database: BoltDB (`habits.db`)

## Key Dependencies
- **Backend**: Chi router, BoltDB, OIDC, Prometheus, Resend, Cobra CLI
- **Frontend**: Vite, TypeScript, TailwindCSS, cal-heatmap

## Quality Checks
Always run before committing:
```bash
make test
```
