# Knowledge Layer — codebase-memory-mcp

The knowledge layer gives Forgia agents deep codebase understanding via
[codebase-memory-mcp](https://github.com/nicobailey/codebase-memory-mcp),
an MCP server that builds a semantic graph of your code (symbols, relations,
call chains).

In Forgia's tiered knowledge architecture this is **Tier 3** — structural
intelligence that goes beyond file contents (Tier 1) and dev-guide conventions
(Tier 2).

## Install

### Binary download

Download the latest release for your platform and place it in your `PATH`:

```bash
# macOS (Apple Silicon)
curl -Lo codebase-memory-mcp \
  https://github.com/nicobailey/codebase-memory-mcp/releases/latest/download/codebase-memory-mcp-darwin-arm64
chmod +x codebase-memory-mcp
sudo mv codebase-memory-mcp /usr/local/bin/
```

### From source (Go >= 1.22)

```bash
go install github.com/nicobailey/codebase-memory-mcp@latest
```

Verify installation:

```bash
command -v codebase-memory-mcp   # should print the path
```

## Setup

Once installed, `forgia init` handles everything automatically:

```bash
forgia init
```

This will:

1. Index the codebase — builds a semantic graph of symbols and relations.
2. Generate `.mcp.json` — configures Claude Code to use the MCP server.

If `.mcp.json` already exists with other servers, the entry is merged (requires
`jq`).

## Verification

```bash
forgia doctor
```

Look for the `codebase-memory-mcp` line:

```
  codebase-memory-mcp       OK    1234 symbols, 5678 edges, synced 5m ago
```

`forgia status` also shows a **Knowledge Layer** section when an index exists:

```
── Knowledge Layer ──
  Tier 3: codebase-memory-mcp  1234 symbols  5678 edges  synced 5m ago
```

## Manual setup

If you prefer not to use `forgia init`, create `.mcp.json` manually in your
project root:

```json
{
  "mcpServers": {
    "codebase-memory-mcp": {
      "type": "stdio",
      "command": "codebase-memory-mcp"
    }
  }
}
```

Then index the repository:

```bash
codebase-memory-mcp cli index_repository path=./
```

## Config reference

The `[knowledge]` section in `.forgia/config.toml` controls the knowledge
layer:

```toml
[knowledge]
provider = "codebase-memory-mcp"
auto_index = true
auto_sync = true
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `provider` | string | `"codebase-memory-mcp"` | Name of the knowledge provider binary. Must be in PATH. |
| `auto_index` | bool | `true` | Run indexing during `forgia init`. Set to `false` to skip. |
| `auto_sync` | bool | `true` | Reserved for future use — live re-indexing on file changes. |

When the `[knowledge]` section is missing entirely, all fields default to the
values shown above.

## Troubleshooting

### Binary not in PATH

```
  codebase-memory-mcp       WARN  not installed (optional)
```

Ensure the binary is in your `PATH`. Check with `command -v codebase-memory-mcp`.

### Stale or empty index

```
  codebase-memory-mcp       WARN  installed but no index (run forgia init)
```

Re-index by running:

```bash
forgia init        # if vault exists, answer y to overwrite prompt
# or directly:
codebase-memory-mcp cli index_repository path=./
```

### Missing .mcp.json

Claude Code won't discover the MCP server without `.mcp.json`. Run
`forgia init` to generate it, or create it manually (see Manual setup above).

### jq not available for merge

If `.mcp.json` already exists and `jq` is not installed, `forgia init` cannot
merge the entry automatically. Install `jq` (`brew install jq`) or add the
entry manually.

### Indexing disabled by config

If `auto_index = false` in `[knowledge]`, `forgia init` will skip indexing.
To re-enable, set `auto_index = true` in `.forgia/config.toml` and re-run
`forgia init`.
