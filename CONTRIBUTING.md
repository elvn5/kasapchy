# Contributing

## Setup

```bash
corepack enable
pnpm install
cp .env.example .env
```

Go 1.27+ is required for `services/`. Fill in `.env` before running anything — see
[README.md](README.md) for details.

## Workflow

1. Create a branch off `main`.
2. Make your changes, following the conventions in [CLAUDE.md](CLAUDE.md) (Clean
   Architecture per Go service, database-per-service, Feature-Sliced Design on the
   frontends, slog for logging, VeeValidate/Zod on the frontend).
3. Before opening a PR, run:
   ```bash
   make lint test build   # Go: gofmt, go vet, tests, build + Biome
   pnpm typecheck
   pnpm build
   ```
4. If you changed a database schema, add a goose migration to
   `services/<name>/migrations/` (e.g. `00002_add_something.sql`). It is applied on
   service start.
5. Open a PR against `main` using the PR template checklist.

## Commit messages

Use a short imperative summary, optionally prefixed with a type (`fix:`, `feat:`,
`chore:`, `refactor:`), e.g. `fix: correct admin JWT expiry`.
