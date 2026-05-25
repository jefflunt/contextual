# Runtime and configuration

## Table of contents
- [Build/runtime](#buildruntime)
- [Config file](#config-file)
- [Config structure](#config-structure)
- [Required settings for Atlassian](#required-settings-for-atlassian)
- [Failure behavior](#failure-behavior)

## Build/runtime

- Go module: `github.com/jefflunt/contextual` (see `go.mod`)
- Scripts (no Makefile):
  - `script/build`              → `go build -o bin/contextual ./cmd/contextual/`
  - `script/test`               → `go test ./...`
  - `script/install`            → installs binary to `$GOPATH/bin`
  - `script/build-test-install` → runs all three in sequence

## Config file

Location: `~/.contextual/config.yml`

The CLI calls `config.Load()` on startup:
- If the file exists and parses: use it.
- If the file does not exist: silently use `config.Config{}`.
- If the file exists but fails to parse: log `[WARN]` and continue with `config.Config{}`.

See `contextual.config.example.yml` in the repo root for a fully annotated example.

## Config structure

```yaml
max_context_length: 10240

atlassian:
  host: your-org.atlassian.net
  api_user: you@example.com
  api_token: your-api-token
  max_spider_jumps: 5
```

Go struct (`internal/config/config.go`):

```go
type Config struct {
    Atlassian        AtlassianConfig `yaml:"atlassian"`
    MaxContextLength int             `yaml:"max_context_length"`
}

type AtlassianConfig struct {
    Host           string `yaml:"host"`
    APIUser        string `yaml:"api_user"`
    APIToken       string `yaml:"api_token"`
    MaxSpiderJumps int    `yaml:"max_spider_jumps"`
}
```

**Note**: `max_context_length` is required.

**No environment variables are used.** All credentials come from the config file.

## Required settings

- `max_context_length`: Required to define the output size for the context file.
- `atlassian.*`: Required if you want to fetch Jira or Confluence items.

If `cfg.Atlassian.Host` is empty:
- Jira fetches are skipped with `[ERROR] atlassian.host not configured`
- Confluence fetches are skipped with the same error

Missing `APIUser`/`APIToken` are not validated at startup but will cause Atlassian API auth failures at fetch time.

## Failure behavior

- Per-arg parse failures: logged `[ERROR]`, item skipped, traversal continues.
- Per-fetch failures: logged `[ERROR]`, item skipped, traversal continues.
- Config load failure: logged `[WARN]`, run continues with empty config.
