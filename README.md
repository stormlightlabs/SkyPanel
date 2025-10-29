# SkyPanel

Bluesky ecosystem: browser extension, self-hosted web client, CLI, and custom feed generator.

## Components

### Browser Extension (`packages/extension/`)

Chrome/Firefox extension providing quick access to Bluesky via sidepanel and newtab.

**Distribution:** Chrome Web Store, Firefox Add-ons, Edge Add-ons

Features:

- Sidepanel and newtab integration
- Timeline, author feeds, and list feeds
- Computed feeds (mutuals, quiet posters)
- Post grouping with read state tracking
- Profile viewing, search
- Built with WXT + Svelte 5 (Runes)

See [extension README](./packages/extension/README.md) for details and [ROADMAP](./packages/extension/ROADMAP.md) for planned features.

### Web Application (`server/`)

**Self-hosted** Bluesky web client and feed generator service.

**Distribution:** Self-deployed via Docker or binary

**Status:** In planning. See [server ROADMAP](./server/ROADMAP.md) for implementation plan.

### CLI (`cli/`)

Terminal interface and command-line tools for Bluesky.

**Distribution:** Binary releases or build from source

Current features:

- Basic status and timeline commands

See [CLI README](./cli/README.md) for details and [ROADMAP](./ROADMAP.md) for planned features.

### Documentation (`packages/docs/`)

User-facing documentation built with Docusaurus.

## Quick Start

### Browser Extension

```sh
pnpm dev:extension

# For Firefox
pnpm --filter @skypanel/extension dev:firefox
```

### CLI

```sh
cd cli
task build
./tmp/skycli --help
```

## Development

### Prerequisites

- Node.js 20+ with `pnpm@9` (run `corepack enable`)
- Go 1.24+
- SQLite 3.x

### Project Structure

```sh
SkyPanel/
├── cli/                    # Go CLI and TUI
├── packages/
│   ├── extension/          # Browser extension
│   └── docs/               # Docusaurus docs
├── server/                 # Go backend + web app
└── scripts/                # Development utilities
```

## Architecture

### Component Relationships

- **Extension** → Distributed via browser stores, works standalone with Bluesky API
- **CLI** → Distributed as binary, interacts with Bluesky
- **Server** → In planning, will be self-hosted (see ROADMAP)

## License

[MIT](https://opensource.org/license/mit)
