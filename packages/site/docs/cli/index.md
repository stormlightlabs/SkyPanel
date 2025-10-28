---
sidebar_position: 1
title: CLI
---

# SkyCLI Overview

SkyCLI is the Bluesky power-user companion for SkyPanel. It gives you direct, scriptable access to the same data model the app relies on: authenticated sessions, feed caches, profile metadata, and exports. The CLI favors explicit output and structured data so you can slot it into pipelines or quick terminal inspections.

## Install & Update

- **From source (recommended during development)**

  ```bash
  go install github.com/stormlightlabs/skypanel/cli/cmd/skycli@latest
  ```

- **Build locally from the repo**

  ```bash
  go build -o skycli ./cli/cmd
  ./skycli --help
  ```

You need Go 1.22+ and a Bluesky app password. SkyCLI persists state in `~/.skycli` (`%APPDATA%\skycli` on Windows):

- `.config.json` — encrypted session metadata
- `cache.db` — SQLite cache populated by fetch/list/search activity

Run `skycli setup` before the first use to create both assets.

## Quick Start

```bash
skycli setup
skycli login --handle @you.bsky.social --password your-app-password
skycli fetch timeline --limit 10
skycli list feeds
```

Common flags are shared across subcommands:

- `--json` (or `-j`) emits raw JSON payloads, matching the underlying API contract.
- `--limit` controls page size for timeline/feed/post retrieval.
- `--cursor` lets you resume pagination using cursors returned in prior responses.

## Command Map

| Command | Purpose |
| --- | --- |
| `setup` | Initialize the config directory and run database migrations. |
| `login` | Authenticate against Bluesky and persist the encrypted session. |
| `status` | Inspect the current session and backend endpoint. |
| `fetch` | Pull timeline, feed, or author posts (writes through to the cache). |
| `list` | List your cached posts or feeds. |
| `search` | Search actors, posts, or locally stored feeds. |
| `view` | Inspect a feed, post, or profile with rich formatting. |
| `export` | Write cached artifacts to disk in JSON, CSV, or TXT formats. |

Each command has a dedicated page with detailed flag coverage and sample output drawn from the Go implementation.
