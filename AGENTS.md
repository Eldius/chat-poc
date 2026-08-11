# AGENTS.md - chat-poc

## Build & Verify

```sh
go build ./...
go test ./... -race
go vet ./...
make validate       # test(-race) + golangci-lint + staticcheck + gocyclo + gosec + govulncheck
```

Env prefix `CHAT`, config file `config.yaml` via `github.com/eldius/initial-config-go`.

### Standards & gates

All new/changed code must pass `make validate` and meet:

| Gate | Threshold |
|------|-----------|
| `gofmt` | clean (non-negotiable) |
| `go vet` / `golangci-lint` / `staticcheck` | clean |
| `gocyclo` (`make cyclo`) | cyclomatic complexity < 10 per function |
| `gosec` (`make sec`) | clean |
| `govulncheck` | no known vulnerabilities |
| `go test -race` | clean |
| Test coverage (`make cover`) | > 80% target for new/changed packages (legacy packages are below; raise on touch) |

Table-driven tests are the default style (testify/assert). Verification tools live in
`go.mod` `tool` directives (`gocyclo`, `staticcheck`, `gosec`, `govulncheck`, `goreleaser`) —
run via `go tool <name>`, no separate install.

## Makefile shortcuts

| Target | Notes |
|--------|-------|
| `make chat` | Requires `DB_USER` / `DB_PASS` env vars. Runs `cmd/cli` `chat` subcommand. |
| `make debug` | `dlv debug --headless --listen=:40237` for `cmd/cli chat --session <uuid>`. |
| `make doc-add` | Ingests URLs into vector store. Edit URLs in Makefile. |
| `make doc-query` | Vector similarity search via `doc query ABECS code`. |
| `make cache-ls` | Lists BoltDB cache entries (`cache ls`). |
| `make snapshot` | `goreleaser release --snapshot --clean` |
| `make cover` | Per-package coverage report (`coverage.out`). |
| `make cyclo` | Fails on functions with cyclomatic complexity ≥ 10. |
| `make sec` | gosec security scan. |
| `make staticcheck` | staticcheck analysis. |
| `make validate` | Runs `test lint staticcheck cyclo sec vulncheck` in sequence. |

Standalone: `golangci-lint run`, `go tool govulncheck ./...`.

## Entrypoint & layout

Single entrypoint: `cmd/cli/` (cobra CLI, `main.go` → `root.go`). Subcommands: `chat`, `doc {add,query}`, `cache ls`.

```
internal/
  cache/         - BoltDB cache.Backend for langchaingo (`.db/cache.db`)
  llm/           - segregated interfaces (Chatter/DocumentStore/CacheLister, composed
                   into Backend) + langchaingo impl (backend.go), ChatCallback
                   (chat_callback.go), client factory (factory.go), otel-metric
                   callback handler (handler.go), ollama.go/openai.go clients
  config/        - viper defaults as setup.Prop vars (constants.go, configs.go)
  tools/docs/    - documentation_search tool (titan-embeddings + chromem)
  tui/chatv2/    - Bubble Tea v2 TUI (model.go, overlay.go, export.go)
```

All subcommands build the backend via `newBackend()` in `cmd/cli/cmd/backend_helper.go`.

## Key constraints

- `ChatScreen(ctx, cb, backendName)` receives its dependencies from `chat.go` (constructor injection); the TUI package does not build LLM clients.
- `doc {add,query}` and `cache ls` call `Backend` methods (`AddDocument`, `QueryDocuments`, `ListCache`) that all return `"not implemented"`.
- Backend `Ask` = plain LLM chat (no tools). `AskWithAgents` = with tools (requires executor). The TUI uses `Ask` only.
- `backend.Name()` reports `opts.Type` (`ollama`/`openai`); only Ollama is exercised in practice.

## TUI keybindings

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+S` | Export chat to `chat-export-*.md` |
| `Ctrl+H` | Toggle help |
| `Esc` | Close popup / quit |
| `Ctrl+C` | Quit |
| Mouse wheel | Scroll chat |

Export saves conversation as Markdown (`chat-export-{timestamp}.md`). The `.gitignore` already covers `chat-export-*.md`.

## Persistence

| Data | Path | Enabled by default? |
|------|------|---------------------|
| Ollama chat history (SQLite) | `chat_cache.db` | Yes (`ollama.generation.cache.enabled: true` in config) |
| `.gitignore` | `*.db` | Yes |
| `.db/` dir | Not in `.gitignore` | Persists after clean |

## Config quirks

- App name for OpenTelemetry: `chat-poc`; goreleaser project name: `ai-chat`, binary: `chat-cli`.
- `CHAT_DB_USER` / `CHAT_DB_PASS` env vars are defined as config props but **not consumed by any code**.
- `--session` flag: defaults to a random UUID on `chat`; explicitly set in `make debug`.

## Tests

Minimal: `internal/cache/boltdb_test.go` (table-driven, testify/assert) and `internal/llm/opts_test.go`. No mocks, no integration tests.

## GoReleaser

```sh
goreleaser release --snapshot --clean
```

Archives include `config.yaml`, `LICENSE`, `README.md`.
