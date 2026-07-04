# AGENTS.md - chat-poc

## Build & Verify

```sh
go build ./...
go test ./...
make validate       # test + golangci-lint + govulncheck
```

Env prefix `CHAT`, config file `config.yaml` via `github.com/eldius/initial-config-go`.

## Makefile shortcuts

| Target | Notes |
|--------|-------|
| `make chat` | Requires `DB_USER` / `DB_PASS` env vars. Runs `cmd/cli` `chat` subcommand. |
| `make debug` | `dlv debug --headless --listen=:40237` for `cmd/cli chat --session <uuid>`. |
| `make doc-add` | Ingests URLs into vector store. Edit URLs in Makefile. |
| `make doc-query` | Vector similarity search via `doc query ABECS code`. |
| `make cache-ls` | Lists BoltDB cache entries (`cache ls`). |
| `make snapshot` | `goreleaser release --snapshot --clean` |
| `make validate` | Runs `test lint vulncheck` in sequence. |

## Entrypoint & layout

Single entrypoint: `cmd/cli/` (cobra CLI). Subcommands: `chat`, `doc {add,query}`, `cache ls`.

```
internal/
  cache/         - BoltDB cache.Backend for langchaingo (`.db/cache.db`)
  client/
    llm/         - Backend interface + otel-metric callback handler
      ollama/    - Ollama LLM backend (only active backend)
  config/        - viper defaults as setup.Prop vars
  service/       - thin ConversationService facade
  tools/docs/    - documentation_search tool (titan-embeddings + chromem)
  tui/chatv2/    - Bubble Tea v2 TUI model
```

## Key constraints

- `ChatScreen()` (in `tui/chatv2/model.go`) directly creates the Ollama backend inline — constructor injection is not used.
- `doc {add,query}` and `cache ls` create an Ollama backend **without tools**; their methods (`AddDocument`, `QueryDocuments`, `ListCache`) are stubs that return `"not implemented"`.

## Persistence

| Data | Path | Enabled by default? |
|------|------|---------------------|
| Ollama chat history (SQLite) | `chat_cache.db` | Yes (`ollama.generation.cache.enabled: true` in config.yaml) |
| `.gitignore` | `*.db` | Yes |
| `.db/` dir | Not in `.gitignore` | Persists after clean |

## Config quirks

- App name for OpenTelemetry: `chat-poc`; goreleaser project name: `ai-chat`, binary: `chat-cli`.
- `CHAT_DB_USER` / `CHAT_DB_PASS` env vars are defined as config props but **not consumed by any code**.
- `--session` flag: defaults to a random UUID on `chat`; explicitly set in `make debug`.

## Tests

Minimal: `internal/cache/boltdb_test.go` (table-driven, testify/assert) and `internal/client/llm/ollama/opts_test.go`. No mocks, no integration tests.

## GoReleaser

```sh
goreleaser release --snapshot --clean
```

Archives include `config.yaml`, `LICENSE`, `README.md`.
