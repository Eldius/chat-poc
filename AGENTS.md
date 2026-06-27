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
| `make debug` | `dlv debug --headless --listen=:40237` for `cmd/cli chat`. |
| `make doc-add` | Ingests URLs into vector store. Edit paths in Makefile. |
| `make doc-query` | Performs vector similarity search. |
| `make cache-ls` | Lists BoltDB cache entries. |
| `make snapshot` | GoReleaser snapshot build. |
| `make clear-log` | Deletes `execution.log`. |

## Architecture

One entrypoint:
- **`cmd/cli/`** - main app: cobra CLI with subcommands `chat`, `doc add/query`, `cache ls`

### Internal layout

```
internal/
  cache/         - BoltDB backend for langchaingo cache
  client/
    llm/         - Backend interface + handler
      ollama/    - Ollama LLM backend
    custom_handler.go - slog-based langchaingo callback handler
  config/        - viper property defaults (constants.go), getters (configs.go)
  service/       - ConversationService facade
  tools/
    docs/        - documentation_search tool (vector store)
  tui/
    chatv2/      - Bubble Tea v2 chat model
```

### LangChain tools

| Tool name | Package | Backend |
|-----------|---------|---------|
| `documentation_search` | `tools/docs` | Chromem vector store with Titan embeddings |

### Persistence (local)

| Data | Path | Enabled by default? |
|------|------|---------------------|
| LLM response cache | `.db/cache.db` (BoltDB) | No (`cache.enabled: false` in config.yaml) |
| Chat memory | `.db/chat.db` (SQLite) | Yes, when `--session` flag is passed |

## Config quirks

- App name: `chat-poc-cli`
- DB credentials passed as `CHAT_DB_USER` / `CHAT_DB_PASS` env vars (not in config.yaml)
- Config defaults live in `internal/config/constants.go` as `setup.Prop` vars
- GoReleaser project name: `ai-chat`, binary: `chat-cli`
- Telemetry disabled by default (`telemetry.enabled: false`)
- `.db/` dir is NOT in `.gitignore` but `*.db` is

## Tests

Minimal coverage:
- `internal/cache/boltdb_test.go` - BoltDB cache Put/Get/List/Override (table-driven, testify/assert)
- `internal/client/llm/ollama/opts_test.go` - Ollama config loading tests

No mocks, no integration tests, no test fixtures.

## GoReleaser

```sh
goreleaser release --snapshot --clean
```

Binary: `chat-cli`, archives include `config.yaml`, `LICENSE`, `README.md`.
