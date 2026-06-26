# AGENTS.md - chat-poc

## Build & Test

```sh
# No linter/typecheck configured. Only verify via:
go build ./...
go test ./...
```

Env prefix `CHAT`, config file `config.yaml` via `github.com/eldius/initial-config-go`.

## Makefile targets

| Command | Notes |
|---------|-------|
| `make chat` | Requires `DB_USER` / `DB_PASS` env vars. Runs `cmd/cli` with `chat` subcommand. |
| `make testing` | TUI dev harness: runs `cmd/testing/` with dummy callback (no AWS, v1 bubbletea). |
| `make testing-v2` | TUI dev harness: runs `cmd/testingv2/` with dummy callback (no AWS, v2 bubbletea). |
| `make debug` | `dlv debug --headless --listen=:40237` for `cmd/cli chat`. |
| `make doc-add` | Ingests URLs into vector store. Edit paths in Makefile. |
| `make doc-query` | Performs vector similarity search. |
| `make cache-ls` | Lists BoltDB cache entries. |
| `make snapshot` | GoReleaser snapshot build. |
| `make clear-log` | Deletes `execution.log`. |

## Architecture

Three entrypoints:
- **`cmd/cli/`** - main app: cobra CLI with subcommands `chat`, `doc add/query`, `cache ls`, `confluence authenticate`
- **`cmd/api/`** - HTTP API stub (not yet implemented; handler is empty)
- **`cmd/testing/`** - TUI test harness (creates Bubble Tea v1 program with fake callback)
- **`cmd/testingv2/`** - TUI test harness (creates Bubble Tea v2 program with fake callback, package `charm.land/bubbletea/v2`)

### Internal layout

```
internal/
  bedrock/       - AWS Bedrock client, agent executor, vector store
  cache/         - BoltDB backend for langchaingo cache
  client/
    bedrock/     - NewBedrockClient factory, options
    confluence/  - OAuth2 client for Confluence
    custom_handler.go - slog-based langchaingo callback handler
  config/        - viper property defaults (constants.go), getters (configs.go)
  service/       - ConversationService, TransactionService facades
  tools/
    docs/        - documentation_search tool (vector store)
    transaction/ - transaction_lookup tool (sqlx + //go:embed SQL queries)
  tui/
    chat/        - Bubble Tea v1 chat model
    chatv2/      - Bubble Tea v2 chat model (newer, parallel implementation)
    confluence/  - OAuth auth flow orchestrator
```

### LangChain tools

| Tool name | Package | Backend |
|-----------|---------|---------|
| `transaction_lookup` | `tools/transaction` | Redshift/Postgres via sqlx, embedded `queries/transaction_lookup.sql` |
| `documentation_search` | `tools/docs` | Chromem vector store with Titan embeddings |

### Persistence (local)

| Data | Path | Enabled by default? |
|------|------|---------------------|
| LLM response cache | `.db/cache.db` (BoltDB) | No (`cache.enabled: false` in config.yaml) |
| Chat memory | `.db/chat.db` (SQLite) | Yes, when `--session` flag is passed |
| Confluence OAuth session | `.db/session.json` | On successful auth |

## Config quirks

- App name for config init differs per binary: `my-chat-app-cli` (CLI) vs `my-chat-app-api` (API)
- DB credentials passed as `CHAT_DB_USER` / `CHAT_DB_PASS` env vars (not in config.yaml)
- Config defaults live in `internal/config/constants.go` as `setup.Prop` vars
- GoReleaser project name: `ai-chat`, binary: `chat-cli`
- Telemetry disabled by default (`telemetry.enabled: false`)
- `.db/` dir is NOT in `.gitignore` but `*.db` is

## Tests

Minimal coverage:
- `internal/tui/chat/functions_test.go` - WordWrap unit tests (table-driven, testify/assert)
- `internal/client/confluence/confluence_test.go` - URL parsing test

No mocks, no integration tests, no test fixtures.

## GoReleaser

```sh
goreleaser release --snapshot --clean
```

Binary: `chat-cli`, archives include `config.yaml`, `LICENSE`, `README.md`.
